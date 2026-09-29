package opportunity

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"learning-orchestrator/backend/internal/contracts"
	"learning-orchestrator/backend/internal/platform"
)

func TestOfferListHonorsEachSourceConsent(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated disposable database")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must name a database ending in _test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	var nonce [8]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	user := "offer-list-" + hex.EncodeToString(nonce[:])
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM opportunity.offers WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	})
	if _, err = db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user); err != nil {
		t.Fatal(err)
	}
	for _, purpose := range []string{"location_context", "calendar_context", "ios_calendar_context", "google_calendar_context"} {
		if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,$2,'1',true,'grant')", user, purpose); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{"ios_geofence", "calendar", "ios_calendar", "google_calendar"} {
		var eventID string
		if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,$2,gen_random_uuid()::text,'resource','practice_place',now(),now()+interval '30 minutes') RETURNING id`, user, source).Scan(&eventID); err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(`INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until)
 VALUES($1,'00000000-0000-4000-8000-000000000001','step',$2,$3,'approved_context_match',now()+interval '30 minutes')`, user, eventID, source); err != nil {
			t.Fatal(err)
		}
	}
	list := func() []Offer {
		t.Helper()
		r := httptest.NewRequest(http.MethodGet, "/internal/offers", nil)
		r.Header.Set("X-User-ID", user)
		w := httptest.NewRecorder()
		(Server{DB: db}).list(w, r)
		if w.Code != 200 {
			t.Fatalf("list status %d: %s", w.Code, w.Body.String())
		}
		var result struct {
			Items []Offer `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result.Items
	}
	if got := list(); len(got) != 4 {
		t.Fatalf("initial offers: %d", len(got))
	}
	if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'google_calendar_context','1',false,'revoke-google')", user); err != nil {
		t.Fatal(err)
	}
	if got := list(); len(got) != 3 {
		t.Fatalf("revoked Google offer remained visible: %+v", got)
	} else {
		for _, offer := range got {
			if offer.ContextLabel == "google_calendar" {
				t.Fatalf("revoked Google offer remained visible: %+v", got)
			}
		}
	}
	if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'location_context','1',false,'revoke')", user); err != nil {
		t.Fatal(err)
	}
	if got := list(); len(got) != 2 {
		t.Fatalf("revoked location offer remained visible: %+v", got)
	}
}

func TestWorkflowWakesAndEvaluates(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	called := false
	env.RegisterActivityWithOptions(func(_ context.Context, in contracts.ContextTrigger) error {
		called = in.EventID == "00000000-0000-4000-8000-000000000001"
		return nil
	}, activity.RegisterOptions{Name: "EvaluateOpportunity"})
	env.ExecuteWorkflow(Workflow, contracts.ContextTrigger{UserID: "user", EventID: "00000000-0000-4000-8000-000000000001"})
	if !env.IsWorkflowCompleted() || env.GetWorkflowError() != nil || !called {
		t.Fatalf("workflow did not evaluate: %v", env.GetWorkflowError())
	}
}

func TestWorkflowWaitsForScheduledContext(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	target := start.Add(48 * time.Hour)
	env.SetStartTime(start)
	var calledAt time.Time
	checkedBefore := false
	env.RegisterDelayedCallback(func() {
		checkedBefore = true
		if !calledAt.IsZero() {
			t.Error("scheduled context evaluated before its time")
		}
	}, 47*time.Hour)
	env.RegisterActivityWithOptions(func(_ context.Context, _ contracts.ContextTrigger) error {
		calledAt = env.Now()
		return nil
	}, activity.RegisterOptions{Name: "EvaluateOpportunity"})
	env.ExecuteWorkflow(Workflow, contracts.ContextTrigger{UserID: "user", EventID: "00000000-0000-4000-8000-000000000001", EvaluateAt: target, DelaySeconds: 2})
	if !env.IsWorkflowCompleted() || env.GetWorkflowError() != nil || !checkedBefore || !calledAt.Equal(target) {
		t.Fatalf("workflow evaluated at %s instead of %s: %v", calledAt, target, env.GetWorkflowError())
	}
}

func TestCancelScheduledWorkflowBeforeEvaluation(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	env.SetStartTime(start)
	called := false
	env.RegisterActivityWithOptions(func(_ context.Context, _ contracts.ContextTrigger) error {
		called = true
		return nil
	}, activity.RegisterOptions{Name: "EvaluateOpportunity"})
	env.RegisterDelayedCallback(env.CancelWorkflow, time.Hour)
	env.ExecuteWorkflow(Workflow, contracts.ContextTrigger{UserID: "user", EventID: "00000000-0000-4000-8000-000000000001", EvaluateAt: start.Add(48 * time.Hour)})
	if !env.IsWorkflowCompleted() || env.GetWorkflowError() == nil || called {
		t.Fatalf("canceled scheduled context evaluated: %v", env.GetWorkflowError())
	}
}

func TestOpportunityBudget(t *testing.T) {
	now := time.Now()
	if mayOffer(1, 0, sql.NullTime{}, now) || mayOffer(0, 2, sql.NullTime{}, now) || mayOffer(0, 1, sql.NullTime{Time: now.Add(-time.Hour), Valid: true}, now) {
		t.Fatal("allowed a duplicate, over-budget, or too-soon offer")
	}
	if !mayOffer(0, 1, sql.NullTime{Time: now.Add(-4 * time.Hour), Valid: true}, now) {
		t.Fatal("suppressed eligible offer")
	}
}

func TestForegroundCalendarFreshness(t *testing.T) {
	now := time.Now()
	if freshForegroundCalendar(now.Add(-6*time.Minute), now) || freshForegroundCalendar(now.Add(2*time.Minute), now) || !freshForegroundCalendar(now.Add(-time.Minute), now) || !foregroundCalendar("google_calendar") || !foregroundCalendar("ios_calendar") {
		t.Fatal("foreground Calendar freshness bounds are wrong")
	}
}

func TestForegroundCalendarOfferRechecksFreshnessAndExpiresQuickly(t *testing.T) {
	for _, source := range []struct{ name, purpose, resource string }{
		{"ios_calendar", "ios_calendar_context", "ios:device"},
		{"google_calendar", "google_calendar_context", "google:device"},
	} {
		t.Run(source.name, func(t *testing.T) { testForegroundCalendarOffer(t, source.name, source.purpose, source.resource) })
	}
}

func testForegroundCalendarOffer(t *testing.T, source, purpose, resource string) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated disposable database")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must name a database ending in _test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	var nonce [8]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	user := "calendar-freshness-" + hex.EncodeToString(nonce[:])
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM opportunity.notification_outbox WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM opportunity.offers WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	})
	if _, err = db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,$2,'1',true,'grant')", user, purpose); err != nil {
		t.Fatal(err)
	}
	var eventID string
	if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until)
 VALUES($1,$2,gen_random_uuid()::text,$3,'calendar_availability',now(),now(),now()+interval '30 minutes') RETURNING id`, user, source, resource).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	var staleDuringCandidate atomic.Bool
	staleDuringCandidate.Store(true)
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/internal/events/"+eventID:
			var observedAt, validUntil time.Time
			if err := db.QueryRow("SELECT observed_at,valid_until FROM context.events WHERE id=$1", eventID).Scan(&observedAt, &validUntil); err != nil {
				t.Error(err)
			}
			_ = json.NewEncoder(w).Encode(contextEvent{ID: eventID, Source: source, Category: "calendar_availability", ObservedAt: observedAt, ValidUntil: validUntil})
		case r.URL.Path == "/internal/opportunity-eligible":
			_ = json.NewEncoder(w).Encode(eligibility{Allowed: true})
		case r.URL.Path == "/internal/candidates":
			if staleDuringCandidate.Load() {
				if _, err := db.Exec("UPDATE context.events SET observed_at=now()-interval '10 minutes' WHERE id=$1", eventID); err != nil {
					t.Error(err)
				}
			}
			_ = json.NewEncoder(w).Encode(candidateResponse{Candidate: &candidate{TrackID: "00000000-0000-4000-8000-000000000001", NodeID: "step", Title: "Try a skill", ReasonCode: "calendar_availability"}})
		default:
			t.Errorf("unexpected request: %s", r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(remote.Close)
	client := platform.Client{Base: remote.URL, HTTP: remote.Client()}
	a := Activities{DB: db, Context: client, Experience: client, Learning: client}
	trigger := contracts.ContextTrigger{UserID: user, EventID: eventID}
	if err := a.Evaluate(context.Background(), trigger); err != nil {
		t.Fatal(err)
	}
	var offers int
	if err := db.QueryRow("SELECT count(*) FROM opportunity.offers WHERE user_id=$1", user).Scan(&offers); err != nil || offers != 0 {
		t.Fatalf("stale scan created an offer: %d %v", offers, err)
	}
	staleDuringCandidate.Store(false)
	if _, err := db.Exec("UPDATE context.events SET observed_at=now() WHERE id=$1", eventID); err != nil {
		t.Fatal(err)
	}
	if err := a.Evaluate(context.Background(), trigger); err != nil {
		t.Fatal(err)
	}
	var validUntil time.Time
	if err := db.QueryRow("SELECT valid_until FROM opportunity.offers WHERE user_id=$1", user).Scan(&validUntil); err != nil {
		t.Fatal(err)
	}
	remaining := time.Until(validUntil)
	if remaining <= 0 || remaining > foregroundCalendarFreshness {
		t.Fatalf("calendar offer outlived fresh scan: %s", remaining)
	}
}

func TestConcurrentCalendarSourcesProduceOneOffer(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated disposable database")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must name a database ending in _test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var nonce [8]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	user := "calendar-dedupe-" + hex.EncodeToString(nonce[:])
	defer func() {
		_, _ = db.Exec("DELETE FROM opportunity.notification_outbox WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM opportunity.offers WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	}()
	if _, err = db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user); err != nil {
		t.Fatal(err)
	}
	for _, purpose := range []string{"ios_calendar_context", "google_calendar_context"} {
		if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,$2,'1',true,'grant')", user, purpose); err != nil {
			t.Fatal(err)
		}
	}
	installation := "calendar-" + hex.EncodeToString(nonce[:])
	if _, err = db.Exec(`INSERT INTO experience.devices(installation_id,user_id,notification_permission,location_permission,calendar_permission,google_calendar_permission)
 VALUES($1,$2,'not_requested','not_requested','enabled','enabled')`, installation, user); err != nil {
		t.Fatal(err)
	}
	events := map[string]string{}
	for _, source := range []string{"ios_calendar", "google_calendar"} {
		var id string
		resource := "ios:" + installation
		if source == "google_calendar" {
			resource = "google:" + installation
		}
		if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until)
 VALUES($1,$2,gen_random_uuid()::text,$3,'calendar_availability',now(),now(),now()+interval '30 minutes') RETURNING id`, user, source, resource).Scan(&id); err != nil {
			t.Fatal(err)
		}
		events[id] = source
	}
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/internal/events/"):
			id := strings.TrimPrefix(r.URL.Path, "/internal/events/")
			_ = json.NewEncoder(w).Encode(contextEvent{ID: id, Source: events[id], Category: "calendar_availability", ObservedAt: time.Now(), ValidUntil: time.Now().Add(30 * time.Minute)})
		case r.URL.Path == "/internal/opportunity-eligible":
			_ = json.NewEncoder(w).Encode(eligibility{Allowed: true})
		case r.URL.Path == "/internal/candidates":
			_ = json.NewEncoder(w).Encode(candidateResponse{Candidate: &candidate{TrackID: "00000000-0000-4000-8000-000000000001", NodeID: "step", Title: "Try a skill", ReasonCode: "calendar_availability"}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer remote.Close()
	client := platform.Client{Base: remote.URL, HTTP: remote.Client()}
	a := Activities{DB: db, Context: client, Experience: client, Learning: client}
	var iosEvent, googleEvent string
	for id, source := range events {
		if source == "ios_calendar" {
			iosEvent = id
		} else {
			googleEvent = id
		}
	}
	if _, err = db.Exec("UPDATE context.events SET observed_at=now()-interval '6 minutes' WHERE id=$1", googleEvent); err != nil {
		t.Fatal(err)
	}
	if err = a.Evaluate(context.Background(), contracts.ContextTrigger{UserID: user, EventID: iosEvent}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = db.QueryRow("SELECT count(*) FROM opportunity.offers WHERE user_id=$1", user).Scan(&count); err != nil || count != 0 {
		t.Fatalf("one stale calendar created an offer: %d %v", count, err)
	}
	if _, err = db.Exec("UPDATE context.events SET observed_at=now(),evaluate_at=now()+interval '15 minutes' WHERE id=$1", googleEvent); err != nil {
		t.Fatal(err)
	}
	if err = a.Evaluate(context.Background(), contracts.ContextTrigger{UserID: user, EventID: iosEvent}); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT count(*) FROM opportunity.offers WHERE user_id=$1", user).Scan(&count); err != nil || count != 0 {
		t.Fatalf("non-overlapping calendars created an offer: %d %v", count, err)
	}
	if _, err = db.Exec("UPDATE context.events SET evaluate_at=now() WHERE id=$1", googleEvent); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errors := make(chan error, len(events))
	for id := range events {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errors <- a.Evaluate(context.Background(), contracts.ContextTrigger{UserID: user, EventID: id})
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err = db.QueryRow("SELECT count(*) FROM opportunity.offers WHERE user_id=$1", user).Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicate calendar opportunity: %d %v", count, err)
	}
	if _, err = db.Exec("UPDATE context.events SET observed_at=now()-interval '6 minutes' WHERE id=$1", googleEvent); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/internal/offers", nil)
	r.Header.Set("X-User-ID", user)
	w := httptest.NewRecorder()
	(Server{DB: db}).list(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"items":[]`) {
		t.Fatalf("stale second source left an offer visible: %d %s", w.Code, w.Body.String())
	}
}

func TestEvaluateRechecksContextAndConsentAfterRemoteRead(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated disposable database")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must name a database ending in _test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	user := "opportunity-race-" + hex.EncodeToString(nonce[:])
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM opportunity.notification_outbox WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM opportunity.offers WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	})
	if _, err := db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key)
 VALUES($1,'location_context','1',true,'grant')`, user); err != nil {
		t.Fatal(err)
	}
	var canceledID, revokedID, pausedID, currentID string
	for _, target := range []*string{&canceledID, &revokedID, &pausedID, &currentID} {
		if err := db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,'ios_geofence',gen_random_uuid()::text,'place','practice_place',now(),now()+interval '30 minutes') RETURNING id`, user).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec("UPDATE context.events SET source='calendar',category='calendar_availability' WHERE id=$1", currentID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key)
 VALUES($1,'calendar_context','1',true,'grant-calendar')`, user); err != nil {
		t.Fatal(err)
	}
	var evaluating string
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/internal/events/"):
			evaluating = strings.TrimPrefix(r.URL.Path, "/internal/events/")
			if strings.TrimPrefix(r.URL.Path, "/internal/events/") == canceledID {
				if _, err := db.Exec("UPDATE context.events SET canceled_at=now() WHERE id=$1", canceledID); err != nil {
					t.Error(err)
				}
			}
			source, category := "ios_geofence", "practice_place"
			if evaluating == currentID {
				source, category = "calendar", "calendar_availability"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": strings.TrimPrefix(r.URL.Path, "/internal/events/"), "source": source, "trackId": "00000000-0000-4000-8000-000000000001", "category": category, "validUntil": time.Now().Add(30 * time.Minute), "canceled": false})
		case r.URL.Path == "/internal/opportunity-eligible":
			_ = json.NewEncoder(w).Encode(map[string]bool{"allowed": true})
		case r.URL.Path == "/internal/candidates":
			if r.URL.Query().Get("trackId") != "00000000-0000-4000-8000-000000000001" {
				t.Error("selected learning path was not forwarded to candidate selection")
			}
			if evaluating == currentID && r.URL.Query().Get("category") != "calendar_availability" {
				t.Error("calendar availability was not forwarded to candidate selection")
			}
			if evaluating == revokedID {
				if _, err := db.Exec(`INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key)
 VALUES($1,'location_context','1',false,'revoke')`, user); err != nil {
					t.Error(err)
				}
			}
			if evaluating == pausedID {
				if _, err := db.Exec("UPDATE experience.profiles SET proactive_paused=true WHERE user_id=$1", user); err != nil {
					t.Error(err)
				}
			}
			reason := "due_review_at_saved_place"
			if evaluating == currentID {
				reason = "review_due_in_calendar_gap"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"candidate": map[string]string{"trackId": "00000000-0000-4000-8000-000000000001", "nodeId": "step", "title": "Try a skill", "reasonCode": reason}})
		default:
			t.Errorf("unexpected request: %s", r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(remote.Close)
	client := platform.Client{Base: remote.URL, HTTP: remote.Client()}
	a := Activities{DB: db, Context: client, Experience: client, Learning: client}
	for _, id := range []string{canceledID, revokedID} {
		if err := a.Evaluate(context.Background(), contracts.ContextTrigger{UserID: user, EventID: id}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key)
 VALUES($1,'location_context','1',true,'restore')`, user); err != nil {
		t.Fatal(err)
	}
	if err := a.Evaluate(context.Background(), contracts.ContextTrigger{UserID: user, EventID: pausedID}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE experience.profiles SET proactive_paused=false WHERE user_id=$1", user); err != nil {
		t.Fatal(err)
	}
	if err := a.Evaluate(context.Background(), contracts.ContextTrigger{UserID: user, EventID: currentID}); err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string]int{canceledID: 0, revokedID: 0, pausedID: 0, currentID: 1} {
		var got int
		if err := db.QueryRow("SELECT count(*) FROM opportunity.offers WHERE context_event_id=$1", id).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("offers for %s: got %d, want %d", id, got, want)
		}
	}
	var reason string
	if err := db.QueryRow("SELECT reason_code FROM opportunity.offers WHERE context_event_id=$1", currentID).Scan(&reason); err != nil || reason != "review_due_in_calendar_gap" {
		t.Fatalf("review explanation was lost: %q, %v", reason, err)
	}
}

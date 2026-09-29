package contextsignal

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"learning-orchestrator/backend/internal/platform"
)

func TestCoordinates(t *testing.T) {
	if !validCoordinates(0, 0) || validCoordinates(91, 0) || validCoordinates(0, -181) || validCoordinates(math.NaN(), 0) {
		t.Fatal("invalid coordinates accepted")
	}
}

func TestPersonalRegionIsolation(t *testing.T) {
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
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("INTERNAL_TOKEN", "context-test-secret-123456")
	var allowed atomic.Bool
	allowed.Store(true)
	experience := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]bool{"allowed": allowed.Load()})
	}))
	t.Cleanup(experience.Close)
	handler := Server{DB: db, Experience: platform.Client{Base: experience.URL, HTTP: experience.Client()}}.Handler()
	var nonce [8]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	owner := "context-test-" + hex.EncodeToString(nonce[:])
	other := owner + "-other"
	installation := owner + "-installation"
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM context.outbox WHERE event_id IN (SELECT id FROM context.events WHERE user_id=$1)", owner)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", owner)
		_, _ = db.Exec("DELETE FROM context.regions WHERE user_id=$1", owner)
		_, _ = db.Exec("DELETE FROM learning.tracks WHERE user_id IN ($1,$2)", owner, other)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", owner)
	})
	var trackID, otherTrackID string
	for user, target := range map[string]*string{owner: &trackID, other: &otherTrackID} {
		if err = db.QueryRow("INSERT INTO learning.tracks(user_id,subject_id,goal) VALUES($1,'spanish','Context test') RETURNING id", user).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", owner); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO experience.devices(installation_id,user_id,notification_permission,location_permission) VALUES($1,$2,'not_requested','enabled')", installation, owner); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'location_context','1',true,'initial')", owner); err != nil {
		t.Fatal(err)
	}
	call := func(user, method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		var data []byte
		if body != nil {
			data, _ = json.Marshal(body)
		}
		r := httptest.NewRequest(method, path, bytes.NewReader(data))
		r.Header.Set("X-Internal-Token", "context-test-secret-123456")
		r.Header.Set("X-User-ID", user)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	assert := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("status %d, want %d: %s", w.Code, status, w.Body.String())
		}
	}
	input := map[string]any{"installationId": installation, "trackId": trackID, "latitude": 37.77, "longitude": -122.42}
	allowed.Store(false)
	assert(call(owner, "POST", "/internal/regions", input), 403)
	allowed.Store(true)
	input["trackId"] = otherTrackID
	assert(call(owner, "POST", "/internal/regions", input), 422)
	input["trackId"] = trackID
	created := call(owner, "POST", "/internal/regions", input)
	assert(created, 201)
	var place struct {
		ID      string `json:"id"`
		Version int    `json:"version"`
		TrackID string `json:"trackId"`
	}
	if err = json.Unmarshal(created.Body.Bytes(), &place); err != nil {
		t.Fatal(err)
	}
	if place.ID == "" || place.Version != 1 || place.TrackID != trackID {
		t.Fatal("place missing ID or version")
	}
	otherList := call(other, "GET", "/internal/regions", nil)
	assert(otherList, 200)
	if strings.Contains(otherList.Body.String(), place.ID) {
		t.Fatal("other account saw private place")
	}
	base := time.Now().UTC().Add(-6 * time.Minute)
	signal := map[string]any{"schemaVersion": 1, "installationId": installation, "clientEventId": "enter-1", "regionId": place.ID, "regionVersion": 1, "transition": "enter", "observedAt": base.Format(time.RFC3339Nano)}
	assert(call(other, "POST", "/internal/signals/location", signal), 403)
	assert(call(owner, "POST", "/internal/signals/location", signal), 202)
	var eventID, storedTrack string
	if err = db.QueryRow("SELECT id,track_id FROM context.events WHERE user_id=$1 AND source_event_id='enter-1'", owner).Scan(&eventID, &storedTrack); err != nil || storedTrack != trackID {
		t.Fatalf("context event lost selected learning path: %q %v", storedTrack, err)
	}
	if event := call(owner, "GET", "/internal/events/"+eventID, nil); event.Code != 200 || !strings.Contains(event.Body.String(), `"trackId":"`+trackID+`"`) {
		t.Fatalf("context API lost selected learning path: %s", event.Body.String())
	}
	assertActive := func(want int) {
		t.Helper()
		var active int
		if err := db.QueryRow("SELECT count(*) FROM context.events WHERE user_id=$1 AND canceled_at IS NULL", owner).Scan(&active); err != nil || active != want {
			t.Fatalf("active context events = %d, want %d: %v", active, want, err)
		}
	}
	newer := map[string]any{"schemaVersion": 1, "installationId": installation, "clientEventId": "enter-2", "regionId": place.ID, "regionVersion": 1, "transition": "enter", "observedAt": base.Add(2 * time.Minute).Format(time.RFC3339Nano)}
	assert(call(owner, "POST", "/internal/signals/location", newer), 202)
	assertActive(2)
	staleExit := map[string]any{"schemaVersion": 1, "installationId": installation, "clientEventId": "exit-1", "regionId": place.ID, "regionVersion": 1, "transition": "exit", "observedAt": base.Add(time.Minute).Format(time.RFC3339Nano)}
	assert(call(owner, "POST", "/internal/signals/location", staleExit), 202)
	assertActive(1)
	latestExit := map[string]any{"schemaVersion": 1, "installationId": installation, "clientEventId": "exit-2", "regionId": place.ID, "regionVersion": 1, "transition": "exit", "observedAt": base.Add(3 * time.Minute).Format(time.RFC3339Nano)}
	assert(call(owner, "POST", "/internal/signals/location", latestExit), 202)
	assertActive(0)
	duplicate := call(owner, "POST", "/internal/signals/location", signal)
	assert(duplicate, 202)
	if !strings.Contains(duplicate.Body.String(), `"duplicate":true`) {
		t.Fatalf("old event replay was not idempotent: %s", duplicate.Body.String())
	}
	staleEnter := map[string]any{"schemaVersion": 1, "installationId": installation, "clientEventId": "enter-stale", "regionId": place.ID, "regionVersion": 1, "transition": "enter", "observedAt": base.Add(150 * time.Second).Format(time.RFC3339Nano)}
	ignored := call(owner, "POST", "/internal/signals/location", staleEnter)
	assert(ignored, 202)
	if !strings.Contains(ignored.Body.String(), `"ignored":true`) {
		t.Fatalf("late enter was not ignored: %s", ignored.Body.String())
	}
	assertActive(0)
	fresh := map[string]any{"schemaVersion": 1, "installationId": installation, "clientEventId": "enter-3", "regionId": place.ID, "regionVersion": 1, "transition": "enter", "observedAt": base.Add(4 * time.Minute).Format(time.RFC3339Nano)}
	assert(call(owner, "POST", "/internal/signals/location", fresh), 202)
	assertActive(1)
	input["latitude"] = 37.78
	updated := call(owner, "POST", "/internal/regions", input)
	assert(updated, 201)
	if err = json.Unmarshal(updated.Body.Bytes(), &place); err != nil {
		t.Fatal(err)
	}
	if place.Version != 2 {
		t.Fatal("place version did not advance")
	}
	assertActive(0)
	assert(call(owner, "POST", "/internal/signals/location", signal), 422)
	fresh["clientEventId"] = "enter-new-version"
	fresh["regionVersion"] = 2
	fresh["observedAt"] = base.Add(5 * time.Minute).Format(time.RFC3339Nano)
	assert(call(owner, "POST", "/internal/signals/location", fresh), 202)
	assertActive(1)
	if _, err = db.Exec("UPDATE context.regions SET enabled=false WHERE id=$1", place.ID); err != nil {
		t.Fatal(err)
	}
	assert(call(owner, "DELETE", "/internal/regions", nil), 200)
	assertActive(0)
	list := call(owner, "GET", "/internal/regions", nil)
	assert(list, 200)
	if strings.Contains(list.Body.String(), place.ID) {
		t.Fatal("removed place still visible")
	}
	assert(call(owner, "DELETE", "/internal/regions", nil), 200)
	var decisions int
	var granted bool
	if err = db.QueryRow(`SELECT count(*) FROM experience.consents
	 WHERE user_id=$1 AND purpose='location_context'`, owner).Scan(&decisions); err != nil || decisions != 2 {
		t.Fatalf("disconnect did not record exactly one revocation: %d %v", decisions, err)
	}
	if err = db.QueryRow(`SELECT granted FROM experience.consents WHERE user_id=$1 AND purpose='location_context'
 ORDER BY decision_seq DESC LIMIT 1`, owner).Scan(&granted); err != nil || granted {
		t.Fatalf("disconnect left location consent active: %v %v", granted, err)
	}
	assert(call(owner, "POST", "/internal/regions", input), 403)
	var scheduledAt time.Time
	if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until)
 VALUES($1,'calendar',gen_random_uuid()::text,'test-calendar','practice_place',now(),now()+interval '2 days',now()+interval '3 days') RETURNING evaluate_at`, owner).Scan(&scheduledAt); err != nil || time.Until(scheduledAt) < 47*time.Hour {
		t.Fatalf("future context was not scheduled: %s %v", scheduledAt, err)
	}
	if _, err = db.Exec(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until)
 VALUES($1,'calendar',gen_random_uuid()::text,'test-calendar','practice_place',now(),now()+interval '2 days',now()+interval '1 day')`, owner); err == nil {
		t.Fatal("context evaluation after expiry was accepted")
	}
	if _, err = db.Exec("UPDATE context.outbox SET state='sent' WHERE event_id IN (SELECT id FROM context.events WHERE user_id=$1)", owner); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for _, item := range []struct {
		observed, valid    time.Time
		evaluate, canceled any
	}{
		{now, now.Add(72 * time.Hour), now.Add(48 * time.Hour), now},
		{now.Add(-2 * time.Hour), now.Add(-time.Hour), nil, nil},
	} {
		var id string
		if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until,canceled_at)
 VALUES($1,'calendar',gen_random_uuid()::text,'test-calendar','practice_place',$2,$3,$4,$5) RETURNING id`, owner, item.observed, item.evaluate, item.valid, item.canceled).Scan(&id); err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec("INSERT INTO context.outbox(event_id,kind) VALUES($1,'observe')", id); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		if err = dispatchOne(context.Background(), db, nil, platform.Client{}); err != nil {
			t.Fatal(err)
		}
	}
	var skipped int
	if err = db.QueryRow(`SELECT count(*) FROM context.outbox o JOIN context.events e ON e.id=o.event_id
 WHERE e.user_id=$1 AND e.source='calendar' AND o.kind='observe' AND o.state='sent' AND (e.canceled_at IS NOT NULL OR e.valid_until<now())`, owner).Scan(&skipped); err != nil || skipped != 2 {
		t.Fatalf("stale observations started a workflow: %d %v", skipped, err)
	}
}

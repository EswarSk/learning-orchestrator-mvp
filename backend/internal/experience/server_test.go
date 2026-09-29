package experience

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"learning-orchestrator/backend/internal/platform"
)

func TestCompletedOpportunityClosesInvitation(t *testing.T) {
	const sessionID = "00000000-0000-4000-8000-000000000001"
	const offerID = "00000000-0000-4000-8000-000000000002"
	state, sourceType, offerStatus := "completed", "opportunity", http.StatusOK
	completed := 0
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/sessions/" + sessionID + "/finish":
			_ = json.NewEncoder(w).Encode(map[string]string{"id": sessionID, "state": state, "sourceType": sourceType, "sourceId": offerID})
		case "/internal/offers/" + offerID + "/complete":
			completed++
			w.WriteHeader(offerStatus)
		default:
			t.Errorf("unexpected service path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer remote.Close()
	client := platform.Client{Base: remote.URL, HTTP: remote.Client()}
	s := Server{Learning: client, Opportunity: client}
	finish := func() *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/finish", bytes.NewBufferString(`{"reason":"completed"}`))
		r.SetPathValue("id", sessionID)
		r.Header.Set("X-User-ID", "learner")
		w := httptest.NewRecorder()
		s.finish(w, r)
		return w
	}
	for i := 1; i <= 2; i++ {
		if w := finish(); w.Code != http.StatusAccepted || completed != i || strings.Contains(w.Body.String(), "sourceId") {
			t.Fatalf("completed opportunity was not closed idempotently: status=%d calls=%d body=%s", w.Code, completed, w.Body.String())
		}
	}
	offerStatus = http.StatusConflict // A canceled invitation must not erase completed learning.
	if w := finish(); w.Code != http.StatusAccepted {
		t.Fatalf("canceled invitation lost session completion: %d %s", w.Code, w.Body.String())
	}
	state = "abandoned"
	if w := finish(); w.Code != http.StatusAccepted || completed != 3 {
		t.Fatalf("abandoned session closed an invitation: %d calls=%d", w.Code, completed)
	}
	state, sourceType = "completed", "curriculum"
	if w := finish(); w.Code != http.StatusAccepted || completed != 3 {
		t.Fatalf("curriculum session closed an invitation: %d calls=%d", w.Code, completed)
	}
}

func TestQuietHoursWrapMidnight(t *testing.T) {
	if !inQuiet("22:30", "21:00", "09:00") || !inQuiet("08:59", "21:00", "09:00") || inQuiet("09:00", "21:00", "09:00") {
		t.Fatal("quiet-hours boundary changed")
	}
	if !inQuiet("13:00", "12:00", "14:00") || inQuiet("15:00", "12:00", "14:00") {
		t.Fatal("daytime quiet-hours boundary changed")
	}
}

func TestSavedPlaceDisconnectRejectsDifferentAccount(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/v1/location/regions?ownerId=account-a", nil)
	r.Header.Set("X-User-ID", "account-b")
	w := httptest.NewRecorder()
	(Server{}).deleteRegions(w, r)
	if w.Code != http.StatusConflict {
		t.Fatalf("different account disconnect status %d: %s", w.Code, w.Body.String())
	}
}

func TestDeletingLastLocationDeviceRevokesSavedPlace(t *testing.T) {
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
	user := "device-location-test-" + hex.EncodeToString(nonce[:])
	region := user + "-region"
	defer func() {
		_, _ = db.Exec("DELETE FROM context.outbox WHERE event_id IN (SELECT id FROM context.events WHERE user_id=$1)", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM context.regions WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM learning.tracks WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	}()
	if _, err = db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user); err != nil {
		t.Fatal(err)
	}
	var track string
	if err = db.QueryRow("INSERT INTO learning.tracks(user_id,subject_id,goal) VALUES($1,'spanish','practice') RETURNING id", user).Scan(&track); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO context.regions(id,area_id,category,latitude,longitude,radius_meters,enabled,user_id,track_id)
 VALUES($1,'personal','practice_place',0,0,150,true,$2,$3)`, region, user, track); err != nil {
		t.Fatal(err)
	}
	for _, device := range []string{"phone-a", "phone-b"} {
		if _, err = db.Exec(`INSERT INTO experience.devices(installation_id,user_id,notification_permission,location_permission,calendar_permission,google_calendar_permission)
 VALUES($1,$2,'denied','enabled','denied','denied')`, user+device, user); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'location_context','1',true,'initial')", user); err != nil {
		t.Fatal(err)
	}
	var event string
	if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,'ios_geofence','entry',$2,'practice_place',now(),now()+interval '20 minutes') RETURNING id`, user, region).Scan(&event); err != nil {
		t.Fatal(err)
	}
	remove := func(device string) {
		t.Helper()
		r := httptest.NewRequest(http.MethodDelete, "/v1/devices/"+user+device, nil)
		r.SetPathValue("installationId", user+device)
		r.Header.Set("X-User-ID", user)
		w := httptest.NewRecorder()
		(Server{DB: db}).deleteDevice(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("remove device: %d %s", w.Code, w.Body.String())
		}
	}
	check := func(enabled, canceled, granted bool, cancelCount int) {
		t.Helper()
		var gotEnabled, gotCanceled, gotGranted bool
		var gotCancelCount int
		if err := db.QueryRow("SELECT enabled FROM context.regions WHERE id=$1", region).Scan(&gotEnabled); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("SELECT canceled_at IS NOT NULL FROM context.events WHERE id=$1", event).Scan(&gotCanceled); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("SELECT granted FROM experience.consents WHERE user_id=$1 AND purpose='location_context' ORDER BY decision_seq DESC LIMIT 1", user).Scan(&gotGranted); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("SELECT count(*) FROM context.outbox WHERE event_id=$1 AND kind='cancel'", event).Scan(&gotCancelCount); err != nil {
			t.Fatal(err)
		}
		if gotEnabled != enabled || gotCanceled != canceled || gotGranted != granted || gotCancelCount != cancelCount {
			t.Fatalf("location state: enabled=%v canceled=%v granted=%v cancels=%d", gotEnabled, gotCanceled, gotGranted, gotCancelCount)
		}
	}
	remove("phone-a")
	check(true, false, true, 0)
	remove("phone-b")
	check(false, true, false, 1)
}

func TestQuietHoursChangeWakesDeferredInvitation(t *testing.T) {
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
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	user := "quiet-test-" + hex.EncodeToString(nonce[:])
	defer func() {
		_, _ = db.Exec("DELETE FROM opportunity.offers WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	}()
	if _, err := db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user); err != nil {
		t.Fatal(err)
	}
	var eventID, offerID string
	if err := db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
	 VALUES($1,'calendar',$1,'test-calendar','practice_place',now(),now()+interval '3 hours') RETURNING id`, user).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until)
	 VALUES($1,gen_random_uuid(),'test',$2,'test','approved_context_match',now()+interval '3 hours') RETURNING id`, user, eventID).Scan(&offerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO opportunity.notification_outbox(offer_id,user_id,next_attempt_at) VALUES($1,$2,now()+interval '1 hour')", offerID, user); err != nil {
		t.Fatal(err)
	}
	call := func(body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(http.MethodPatch, "/v1/me", bytes.NewBufferString(body))
		r.Header.Set("X-User-ID", user)
		w := httptest.NewRecorder()
		(Server{DB: db}).patchMe(w, r)
		return w
	}
	if w := call(`{"quietHours":{"start":"25:00","end":"08:00"}}`); w.Code != 422 {
		t.Fatalf("invalid quiet hours status %d: %s", w.Code, w.Body.String())
	}
	if w := call(`{"quietHours":{"start":"22:00","end":"08:00"}}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"quietStart":"22:00"`) {
		t.Fatalf("quiet hours update status %d: %s", w.Code, w.Body.String())
	}
	var next time.Time
	if err := db.QueryRow("SELECT next_attempt_at FROM opportunity.notification_outbox WHERE offer_id=$1", offerID).Scan(&next); err != nil || next.After(time.Now().Add(10*time.Second)) {
		t.Fatalf("deferred invitation was not reconsidered: %s %v", next, err)
	}
}

func TestContextConsentRevocationIsSourceSpecific(t *testing.T) {
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
	user := "consent-test-" + hex.EncodeToString(nonce[:])
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM context.outbox WHERE event_id IN (SELECT id FROM context.events WHERE user_id=$1)", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	})
	if _, err = db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user); err != nil {
		t.Fatal(err)
	}
	for _, purpose := range []string{"location_context", "calendar_context"} {
		if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,$2,'1',true,'grant')", user, purpose); err != nil {
			t.Fatal(err)
		}
	}
	var locationID, calendarID string
	for source, target := range map[string]*string{"ios_geofence": &locationID, "calendar": &calendarID} {
		if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,$2,gen_random_uuid()::text,'test-resource','practice_place',now(),now()+interval '30 minutes') RETURNING id`, user, source).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	s := Server{DB: db}
	check := func(source string, want bool) {
		t.Helper()
		r := httptest.NewRequest(http.MethodGet, "/internal/opportunity-eligible?source="+source, nil)
		r.Header.Set("X-User-ID", user)
		allowed, err := s.proactiveAllowed(r, source)
		if err != nil || allowed != want {
			t.Fatalf("source %q allowed=%v, want %v: %v", source, allowed, want, err)
		}
	}
	check("ios_geofence", true)
	check("calendar", true)
	check("unknown", false)
	r := httptest.NewRequest(http.MethodPost, "/v1/consents", bytes.NewBufferString(`{"purpose":"location_context","policyVersion":"1","granted":false}`))
	r.Header.Set("X-User-ID", user)
	r.Header.Set("Idempotency-Key", "revoke-location")
	w := httptest.NewRecorder()
	s.consent(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("revoke status %d: %s", w.Code, w.Body.String())
	}
	check("ios_geofence", false)
	check("calendar", true)
	for _, test := range []struct {
		id       string
		canceled bool
	}{{locationID, true}, {calendarID, false}} {
		var canceled bool
		if err = db.QueryRow("SELECT canceled_at IS NOT NULL FROM context.events WHERE id=$1", test.id).Scan(&canceled); err != nil || canceled != test.canceled {
			t.Fatalf("event %s canceled=%v, want %v: %v", test.id, canceled, test.canceled, err)
		}
	}
	var cancels int
	if err = db.QueryRow("SELECT count(*) FROM context.outbox WHERE event_id=$1 AND kind='cancel'", locationID).Scan(&cancels); err != nil || cancels != 1 {
		t.Fatalf("cancel work count %d: %v", cancels, err)
	}
}

func TestDeviceRegistrationPreservesPushAndIsolatesAccounts(t *testing.T) {
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
	owner := "device-test-" + hex.EncodeToString(nonce[:])
	other := owner + "-other"
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM experience.devices WHERE user_id IN ($1,$2)", owner, other)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id IN ($1,$2)", owner, other)
	})
	s := Server{DB: db}
	call := func(user, id, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("PUT", "/v1/devices/"+id, bytes.NewBufferString(body))
		r.SetPathValue("installationId", id)
		r.Header.Set("X-User-ID", user)
		w := httptest.NewRecorder()
		s.device(w, r)
		return w
	}
	assert := func(w *httptest.ResponseRecorder, want int) {
		t.Helper()
		if w.Code != want {
			t.Fatalf("status %d, want %d: %s", w.Code, want, w.Body.String())
		}
	}
	id := owner + "-device"
	assert(call(owner, id, `{"pushToken":"ExpoPushToken[device-test]","notificationPermission":"granted"}`), 200)
	assert(call(owner, id, `{"locationPermission":"enabled"}`), 200)
	var token, permission, location string
	if err := db.QueryRow("SELECT push_token,notification_permission,location_permission FROM experience.devices WHERE installation_id=$1", id).Scan(&token, &permission, &location); err != nil {
		t.Fatal(err)
	}
	if token != "ExpoPushToken[device-test]" || permission != "granted" || location != "enabled" {
		t.Fatalf("location enrollment overwrote push: %q %q %q", token, permission, location)
	}
	assert(call(other, id, `{"locationPermission":"enabled"}`), 409)
	assert(call(other, other+"-device", `{"pushToken":"ExpoPushToken[device-test]","notificationPermission":"not_requested"}`), 200)
	var oldToken sql.NullString
	if err := db.QueryRow("SELECT push_token FROM experience.devices WHERE installation_id=$1", id).Scan(&oldToken); err != nil {
		t.Fatal(err)
	}
	if oldToken.Valid {
		t.Fatal("push token remained bound to former account")
	}
	assert(call(other, other+"-device", `{"pushToken":null,"notificationPermission":"denied"}`), 200)
	if err := db.QueryRow("SELECT push_token FROM experience.devices WHERE installation_id=$1", other+"-device").Scan(&oldToken); err != nil {
		t.Fatal(err)
	}
	if oldToken.Valid {
		t.Fatal("explicit token revocation failed")
	}
}

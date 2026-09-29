package experience

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestCalendarDeviceRevocationIsSourceSpecific(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated disposable database")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must end in _test")
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
	user := "calendar-device-test-" + hex.EncodeToString(nonce[:])
	device := user + "-device"
	defer func() {
		_, _ = db.Exec("DELETE FROM context.outbox WHERE event_id IN (SELECT id FROM context.events WHERE user_id=$1)", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	}()
	if _, err = db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO experience.devices(installation_id,user_id,notification_permission,location_permission,calendar_permission,google_calendar_permission) VALUES($1,$2,'denied','denied','enabled','enabled')", device, user); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'ios_calendar_context','1',true,'grant')", user); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'google_calendar_context','1',true,'grant')", user); err != nil {
		t.Fatal(err)
	}
	var event string
	if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until)
 VALUES($1,'ios_calendar',gen_random_uuid()::text,$2,'calendar_availability',now(),now()+interval '20 minutes',now()+interval '1 hour') RETURNING id`, user, "ios:"+device).Scan(&event); err != nil {
		t.Fatal(err)
	}
	var googleEvent string
	if err = db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,evaluate_at,valid_until)
 VALUES($1,'google_calendar',gen_random_uuid()::text,$2,'calendar_availability',now(),now()+interval '20 minutes',now()+interval '1 hour') RETURNING id`, user, "google:"+device).Scan(&googleEvent); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("PUT", "/v1/devices/"+device, bytes.NewBufferString(`{"calendarPermission":"denied"}`))
	r.SetPathValue("installationId", device)
	r.Header.Set("X-User-ID", user)
	w := httptest.NewRecorder()
	(Server{DB: db}).device(w, r)
	if w.Code != 200 {
		t.Fatalf("device denial: %d %s", w.Code, w.Body.String())
	}
	var canceled, granted bool
	if err = db.QueryRow("SELECT canceled_at IS NOT NULL FROM context.events WHERE id=$1", event).Scan(&canceled); err != nil || !canceled {
		t.Fatalf("event not canceled: %v %v", canceled, err)
	}
	if err = db.QueryRow("SELECT granted FROM experience.consents WHERE user_id=$1 AND purpose='ios_calendar_context' ORDER BY decision_seq DESC LIMIT 1", user).Scan(&granted); err != nil || granted {
		t.Fatalf("consent not revoked: %v %v", granted, err)
	}
	var cancelCount int
	if err = db.QueryRow("SELECT count(*) FROM context.outbox WHERE event_id=$1 AND kind='cancel'", event).Scan(&cancelCount); err != nil || cancelCount != 1 {
		t.Fatalf("cancel not queued: %d %v", cancelCount, err)
	}
	if err = db.QueryRow("SELECT canceled_at IS NOT NULL FROM context.events WHERE id=$1", googleEvent).Scan(&canceled); err != nil || canceled {
		t.Fatalf("iPhone revocation canceled Google: %v %v", canceled, err)
	}
	r = httptest.NewRequest("PUT", "/v1/devices/"+device, bytes.NewBufferString(`{"googleCalendarPermission":"denied"}`))
	r.SetPathValue("installationId", device)
	r.Header.Set("X-User-ID", user)
	w = httptest.NewRecorder()
	(Server{DB: db}).device(w, r)
	if w.Code != 200 {
		t.Fatalf("Google denial: %d %s", w.Code, w.Body.String())
	}
	if err = db.QueryRow("SELECT canceled_at IS NOT NULL FROM context.events WHERE id=$1", googleEvent).Scan(&canceled); err != nil || !canceled {
		t.Fatalf("Google event not canceled: %v %v", canceled, err)
	}
	if err = db.QueryRow("SELECT granted FROM experience.consents WHERE user_id=$1 AND purpose='google_calendar_context' ORDER BY decision_seq DESC LIMIT 1", user).Scan(&granted); err != nil || granted {
		t.Fatalf("Google consent not revoked: %v %v", granted, err)
	}
}

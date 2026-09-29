package contextsignal

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestCalendarSnapshots(t *testing.T) {
	for _, source := range []calendarSource{iosCalendar, googleCalendar} {
		t.Run(source.name, func(t *testing.T) { testCalendarSnapshots(t, source) })
	}
}

func testCalendarSnapshots(t *testing.T, source calendarSource) {
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
	owner := "calendar-test-" + hex.EncodeToString(nonce[:])
	other, installation := owner+"-other", owner+"-device"
	defer func() {
		_, _ = db.Exec("DELETE FROM context.outbox WHERE event_id IN (SELECT id FROM context.events WHERE user_id=$1)", owner)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", owner)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", owner)
	}()
	if _, err = db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", owner); err != nil {
		t.Fatal(err)
	}
	iosPermission, googlePermission := "denied", "denied"
	if source.name == "ios_calendar" {
		iosPermission = "enabled"
	} else {
		googlePermission = "enabled"
	}
	if _, err = db.Exec(`INSERT INTO experience.devices(installation_id,user_id,notification_permission,location_permission,calendar_permission,google_calendar_permission)
 VALUES($1,$2,'denied','denied',$3,$4)`, installation, owner, iosPermission, googlePermission); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,$2,'1',true,'grant')", owner, source.purpose); err != nil {
		t.Fatal(err)
	}
	t.Setenv("INTERNAL_TOKEN", "calendar-test-secret-123456")
	handler := (Server{DB: db}).Handler()
	call := func(user, method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		data, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(data))
		r.Header.Set("X-Internal-Token", "calendar-test-secret-123456")
		r.Header.Set("X-User-ID", user)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	start := time.Now().UTC().Add(2 * time.Minute).Truncate(time.Second)
	window := func(start time.Time) map[string]any {
		return map[string]any{"start": start, "end": start.Add(45 * time.Minute)}
	}
	input := map[string]any{"installationId": installation, "windows": []any{window(start)}}
	path := "/internal/calendar/" + strings.TrimSuffix(source.prefix, ":") + "/windows"
	otherPath := "/internal/calendar/ios/windows"
	if source.name == "ios_calendar" {
		otherPath = "/internal/calendar/google/windows"
	}
	if w := call(owner, "POST", otherPath, input); w.Code != 403 {
		t.Fatalf("other calendar permission accepted: %d %s", w.Code, w.Body.String())
	}
	if w := call(owner, "POST", path, map[string]any{"installationId": installation, "windows": []any{window(start.Add(time.Hour))}}); w.Code != 422 {
		t.Fatalf("future iPhone window accepted: %d %s", w.Code, w.Body.String())
	}
	if w := call(owner, "POST", path, map[string]any{"installationId": installation, "windows": []any{window(start), window(start.Add(45 * time.Minute))}}); w.Code != 422 {
		t.Fatalf("multiple iPhone windows accepted: %d %s", w.Code, w.Body.String())
	}
	if w := call(other, "POST", path, input); w.Code != 403 {
		t.Fatalf("cross-account snapshot: %d %s", w.Code, w.Body.String())
	}
	if w := call(owner, "POST", path, input); w.Code != 202 {
		t.Fatalf("first snapshot: %d %s", w.Code, w.Body.String())
	}
	var originalID string
	if err = db.QueryRow("SELECT id FROM context.events WHERE user_id=$1 AND source=$2", owner, source.name).Scan(&originalID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE context.events SET observed_at=now()-interval '10 minutes' WHERE id=$1", originalID); err != nil {
		t.Fatal(err)
	}
	if w := call(owner, "POST", path, input); w.Code != 202 {
		t.Fatalf("repeat snapshot: %d %s", w.Code, w.Body.String())
	}
	var refreshedAt time.Time
	if err = db.QueryRow("SELECT observed_at FROM context.events WHERE id=$1", originalID).Scan(&refreshedAt); err != nil || time.Since(refreshedAt) > time.Minute {
		t.Fatalf("unchanged snapshot did not refresh observation: %s %v", refreshedAt, err)
	}
	var active, observes, cancels int
	count := func() {
		t.Helper()
		if err := db.QueryRow(`SELECT count(*) FILTER (WHERE canceled_at IS NULL) FROM context.events WHERE user_id=$1 AND source=$2`, owner, source.name).Scan(&active); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow(`SELECT count(*) FILTER (WHERE kind='observe'),count(*) FILTER (WHERE kind='cancel')
 FROM context.outbox WHERE event_id IN (SELECT id FROM context.events WHERE user_id=$1)`, owner).Scan(&observes, &cancels); err != nil {
			t.Fatal(err)
		}
	}
	count()
	if active != 1 || observes != 1 || cancels != 0 {
		t.Fatalf("repeat duplicated event: active=%d observe=%d cancel=%d", active, observes, cancels)
	}
	input["windows"] = []any{window(start.Add(30 * time.Second))}
	if w := call(owner, "POST", path, input); w.Code != 202 {
		t.Fatalf("changed snapshot: %d %s", w.Code, w.Body.String())
	}
	count()
	if active != 1 || observes != 2 || cancels != 1 {
		t.Fatalf("edit was not reconciled: active=%d observe=%d cancel=%d", active, observes, cancels)
	}
	if w := call(owner, "DELETE", path+"?installationId="+url.QueryEscape(installation), nil); w.Code != 200 {
		t.Fatalf("disconnect: %d %s", w.Code, w.Body.String())
	}
	count()
	if active != 0 || cancels != 2 {
		t.Fatalf("disconnect left windows: active=%d cancel=%d", active, cancels)
	}
	var granted bool
	if err = db.QueryRow(`SELECT granted FROM experience.consents WHERE user_id=$1 AND purpose=$2
 ORDER BY decision_seq DESC LIMIT 1`, owner, source.purpose).Scan(&granted); err != nil || granted {
		t.Fatalf("disconnect left consent active: %v %v", granted, err)
	}
	if w := call(owner, "POST", path, input); w.Code != 403 {
		t.Fatalf("disconnected device accepted snapshot: %d %s", w.Code, w.Body.String())
	}
}

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
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNotificationDeliveryGuards(t *testing.T) {
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
	user := "push-test-" + hex.EncodeToString(nonce[:])
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM opportunity.notification_outbox WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM opportunity.offers WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.consents WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.devices WHERE user_id=$1", user)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", user)
	})
	must := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	must("INSERT INTO experience.profiles(user_id,quiet_start,quiet_end) VALUES($1,'23:00','23:00')", user)
	must("INSERT INTO experience.devices(installation_id,user_id,push_token,notification_permission,location_permission) VALUES($1,$2,'ExpoPushToken[test]','granted','enabled')", user+"-device", user)
	must("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'location_context','1',true,'location')", user)
	must("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'notifications','1',true,'push')", user)
	newOffer := func() string {
		t.Helper()
		var event, offer string
		if err := db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,'ios_geofence',gen_random_uuid()::text,'place','practice_place',now(),now()+interval '30 minutes') RETURNING id`, user).Scan(&event); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow(`INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until)
 VALUES($1,gen_random_uuid(),'step',$2,'Private skill title','saved_practice_place',now()+interval '30 minutes') RETURNING id`, user, event).Scan(&offer); err != nil {
			t.Fatal(err)
		}
		must("INSERT INTO opportunity.notification_outbox(offer_id,user_id) VALUES($1,$2)", offer, user)
		return offer
	}
	var calls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var message struct {
			To, Title, Body string
			Data            map[string]string
		}
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			t.Error(err)
		}
		if message.To != "ExpoPushToken[test]" || strings.Contains(message.Title+message.Body, "Private skill title") || message.Data["offerId"] == "" {
			t.Errorf("unsafe push payload: %+v", message)
		}
		_, _ = w.Write([]byte(`{"data":{"status":"ok","id":"ticket-test"}}`))
	}))
	t.Cleanup(provider.Close)
	checkState := func(offer, want string) {
		t.Helper()
		var got string
		if err := db.QueryRow("SELECT state FROM opportunity.notification_outbox WHERE offer_id=$1", offer).Scan(&got); err != nil || got != want {
			t.Fatalf("state %q, want %q: %v", got, want, err)
		}
	}
	ctx := context.Background()

	first := newOffer()
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(first, "sent")
	if calls.Load() != 1 {
		t.Fatal("eligible offer was not sent")
	}
	receipt := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			IDs []string `json:"ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if len(request.IDs) != 1 || request.IDs[0] != "ticket-test" {
			t.Errorf("wrong receipt request: %+v", request)
		}
		_, _ = w.Write([]byte(`{"data":{"ticket-test":{"status":"ok"}}}`))
	}))
	t.Cleanup(receipt.Close)
	must("UPDATE opportunity.notification_outbox SET receipt_next_check_at=now() WHERE offer_id=$1", first)
	if err := checkReceipts(ctx, db, receipt.Client(), receipt.URL); err != nil {
		t.Fatal(err)
	}
	var receiptState string
	if err := db.QueryRow("SELECT receipt_status FROM opportunity.notification_outbox WHERE offer_id=$1", first).Scan(&receiptState); err != nil || receiptState != "accepted" {
		t.Fatalf("APNs receipt not recorded: %q %v", receiptState, err)
	}
	var retained int
	if err := db.QueryRow("SELECT count(*) FROM opportunity.notification_outbox WHERE offer_id=$1 AND (push_token_hash IS NOT NULL OR device_installation_id IS NOT NULL)", first).Scan(&retained); err != nil || retained != 0 {
		t.Fatalf("completed receipt retained device linkage: %d %v", retained, err)
	}

	second := newOffer()
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(second, "suppressed") // One notification per local day.

	third := newOffer()
	must("UPDATE opportunity.notification_outbox SET sent_at=now()-interval '2 days' WHERE offer_id=$1", first)
	must("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'notifications','1',false,'revoke')", user)
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(third, "suppressed")
	if calls.Load() != 1 {
		t.Fatal("sent after revocation")
	}

	fourth := newOffer()
	must("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'notifications','1',true,'restore')", user)
	must("UPDATE context.events SET canceled_at=now() WHERE id=(SELECT context_event_id FROM opportunity.offers WHERE id=$1)", fourth)
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(fourth, "suppressed")
	if calls.Load() != 1 {
		t.Fatal("sent after context cancellation")
	}

	fifth := newOffer()
	nowUTC := time.Now().UTC()
	must("UPDATE experience.profiles SET quiet_start=$2::time,quiet_end=$3::time,timezone='UTC' WHERE user_id=$1", user, nowUTC.Add(-time.Minute).Format("15:04"), nowUTC.Add(2*time.Minute).Format("15:04"))
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(fifth, "pending")
	var wake time.Time
	if err := db.QueryRow("SELECT next_attempt_at FROM opportunity.notification_outbox WHERE offer_id=$1", fifth).Scan(&wake); err != nil || !wake.After(nowUTC) || wake.After(nowUTC.Add(3*time.Minute)) {
		t.Fatalf("quiet-hour wake was not deferred to the next eligible minute: %v, %v", wake, err)
	}
	must("UPDATE experience.profiles SET quiet_start='23:00',quiet_end='23:00' WHERE user_id=$1", user)
	must("UPDATE opportunity.notification_outbox SET next_attempt_at=now() WHERE offer_id=$1", fifth)
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(fifth, "sent")
	must("UPDATE opportunity.notification_outbox SET sent_at=now()-interval '2 days' WHERE offer_id=$1", fifth)
	sixth := newOffer()
	var retries atomic.Int32
	flaky := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if retries.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"status":"ok","id":"ticket-after-retry"}}`))
	}))
	t.Cleanup(flaky.Close)
	if err := dispatchNotification(ctx, db, flaky.Client(), flaky.URL); err != nil {
		t.Fatal(err)
	}
	checkState(sixth, "pending")
	must("UPDATE opportunity.notification_outbox SET next_attempt_at=now() WHERE offer_id=$1", sixth)
	if err := dispatchNotification(ctx, db, flaky.Client(), flaky.URL); err != nil {
		t.Fatal(err)
	}
	checkState(sixth, "sent")
	if retries.Load() != 2 {
		t.Fatal("retry did not complete exactly once after provider recovery")
	}
	must("UPDATE opportunity.notification_outbox SET sent_at=now()-interval '2 days' WHERE offer_id=$1", sixth)
	seventh := newOffer()
	longQuiet := time.Now().UTC()
	must("UPDATE experience.profiles SET quiet_start=$2::time,quiet_end=$3::time WHERE user_id=$1", user,
		longQuiet.Add(-time.Minute).Format("15:04"), longQuiet.Add(time.Hour).Format("15:04"))
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(seventh, "suppressed")
	must("UPDATE experience.profiles SET quiet_start='23:00',quiet_end='23:00' WHERE user_id=$1", user)
	eighth := newOffer()
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(eighth, "sent")
	invalid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"ticket-test":{"status":"error","details":{"error":"DeviceNotRegistered"}}}}`))
	}))
	t.Cleanup(invalid.Close)
	must("UPDATE opportunity.notification_outbox SET receipt_next_check_at=now() WHERE offer_id=$1", eighth)
	if err := checkReceipts(ctx, db, invalid.Client(), invalid.URL); err != nil {
		t.Fatal(err)
	}
	var token sql.NullString
	if err := db.QueryRow("SELECT push_token FROM experience.devices WHERE installation_id=$1", user+"-device").Scan(&token); err != nil || token.Valid {
		t.Fatalf("invalid Expo token still registered: %v %v", token, err)
	}
	if err := db.QueryRow("SELECT receipt_status FROM opportunity.notification_outbox WHERE offer_id=$1", eighth).Scan(&receiptState); err != nil || receiptState != "rejected" {
		t.Fatalf("rejected receipt not recorded: %q %v", receiptState, err)
	}
	must("UPDATE experience.devices SET push_token='ExpoPushToken[replacement]',notification_permission='granted' WHERE installation_id=$1", user+"-device")
	stale := newOffer()
	must(`UPDATE opportunity.notification_outbox SET state='sent',provider_ticket='ticket-stale',sent_at=now(),
 device_installation_id=$2,push_token_hash=encode(digest('ExpoPushToken[test]','sha256'),'hex'),
 receipt_status='pending',receipt_next_check_at=now() WHERE offer_id=$1`, stale, user+"-device")
	staleReceipt := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"ticket-stale":{"status":"error","details":{"error":"DeviceNotRegistered"}}}}`))
	}))
	t.Cleanup(staleReceipt.Close)
	if err := checkReceipts(ctx, db, staleReceipt.Client(), staleReceipt.URL); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT push_token FROM experience.devices WHERE installation_id=$1", user+"-device").Scan(&token); err != nil || !token.Valid || token.String != "ExpoPushToken[replacement]" {
		t.Fatalf("stale receipt removed replacement token: %v %v", token, err)
	}
	ninth := newOffer()
	must(`UPDATE opportunity.notification_outbox SET state='sent',provider_ticket='ticket-missing',sent_at=now(),
 receipt_status='pending',receipt_next_check_at=now() WHERE offer_id=$1`, ninth)
	missing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	t.Cleanup(missing.Close)
	for attempt := 0; attempt < 3; attempt++ {
		if err := checkReceipts(ctx, db, missing.Client(), missing.URL); err != nil {
			t.Fatal(err)
		}
		if attempt < 2 {
			must("UPDATE opportunity.notification_outbox SET receipt_next_check_at=now() WHERE offer_id=$1", ninth)
		}
	}
	if err := db.QueryRow("SELECT receipt_status FROM opportunity.notification_outbox WHERE offer_id=$1", ninth).Scan(&receiptState); err != nil || receiptState != "missing" {
		t.Fatalf("missing receipt was not surfaced: %q %v", receiptState, err)
	}
	ios := newOffer()
	must("UPDATE context.events SET source='ios_calendar',category='calendar_availability' WHERE id=(SELECT context_event_id FROM opportunity.offers WHERE id=$1)", ios)
	must("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'ios_calendar_context','1',true,'ios')", user)
	before := calls.Load()
	if err := dispatchNotification(ctx, db, provider.Client(), provider.URL); err != nil {
		t.Fatal(err)
	}
	checkState(ios, "suppressed") // Foreground-only calendar sync cannot justify unattended push.
	if calls.Load() != before {
		t.Fatal("iPhone Calendar caused an unattended push")
	}
}

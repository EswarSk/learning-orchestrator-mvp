package experience

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAccountExportIsCompleteAndOwnerScoped(t *testing.T) {
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
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	owner := "export-test-" + hex.EncodeToString(nonce[:])
	other := owner + "-other"
	t.Cleanup(func() {
		for _, query := range []string{
			"DELETE FROM opportunity.notification_outbox WHERE user_id IN ($1,$2)",
			"DELETE FROM opportunity.offers WHERE user_id IN ($1,$2)",
			"DELETE FROM context.outbox WHERE event_id IN (SELECT id FROM context.events WHERE user_id IN ($1,$2))",
			"DELETE FROM context.events WHERE user_id IN ($1,$2)",
			"DELETE FROM context.regions WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.evidence WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.sessions WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.skill_state WHERE track_id IN (SELECT id FROM learning.tracks WHERE user_id IN ($1,$2))",
			"DELETE FROM learning.progress WHERE track_id IN (SELECT id FROM learning.tracks WHERE user_id IN ($1,$2))",
			"DELETE FROM learning.track_content WHERE track_id IN (SELECT id FROM learning.tracks WHERE user_id IN ($1,$2))",
			"DELETE FROM learning.idempotency WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.course_generation_quota WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.tracks WHERE user_id IN ($1,$2)",
			"DELETE FROM experience.profiles WHERE user_id IN ($1,$2)",
		} {
			if _, err := db.Exec(query, owner, other); err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
	})
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, user := range []string{owner, other} {
		exec("INSERT INTO experience.profiles(user_id) VALUES($1)", user)
		exec("INSERT INTO experience.devices(installation_id,user_id,push_token,notification_permission,location_permission) VALUES($1,$2,$3,'granted','enabled')", user+"-device", user, "ExpoPushToken["+user+"]")
		exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'location_context','v1',true,$2)", user, "private-key-"+user)
		var track, session, event, offer string
		if err := db.QueryRow("INSERT INTO learning.tracks(user_id,subject_id,goal) VALUES($1,'guitar',$2) RETURNING id", user, "goal-"+user).Scan(&track); err != nil {
			t.Fatal(err)
		}
		exec("INSERT INTO learning.track_content(track_id,manifest,model) VALUES($1,$2,'test')", track, `{"subject":{"title":"Guitar"}}`)
		exec("INSERT INTO learning.progress(track_id,node_id) VALUES($1,'node-1')", track)
		if err := db.QueryRow("INSERT INTO learning.sessions(user_id,track_id,node_id,source_type,source_id,mode) VALUES($1,$2,'node-1','curriculum','node-1','text') RETURNING id", user, track).Scan(&session); err != nil {
			t.Fatal(err)
		}
		exec("INSERT INTO learning.turns(session_id,message_id,learner_text,tutor_text) VALUES($1,'turn-1',$2,'Try again')", session, "learner answer "+user)
		exec("INSERT INTO learning.evidence(user_id,track_id,skill_id,session_id,kind,outcome,assistance) VALUES($1,$2,'skill-1',$3,'practice','applied','none')", user, track, session)
		exec("INSERT INTO learning.skill_state(track_id,skill_id,stage,next_review_at) VALUES($1,'skill-1','practicing',now())", track)
		exec("INSERT INTO learning.course_generation_quota(user_id,day,attempts) VALUES($1,current_date,1)", user)
		exec("INSERT INTO learning.idempotency(user_id,operation,request_key,request_hash,response) VALUES($1,'test','internal-key','internal-hash',$2)", user, `{"private":"internal-response"}`)
		exec("INSERT INTO context.regions(id,area_id,category,latitude,longitude,radius_meters,enabled,user_id) VALUES($1,'home','practice_place',37,-122,100,true,$2)", user+"-region", user)
		if err := db.QueryRow("INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until) VALUES($1,'location',$2,$3,'practice_place',now(),now()+interval '1 hour') RETURNING id", user, user+"-event", user+"-region").Scan(&event); err != nil {
			t.Fatal(err)
		}
		exec("INSERT INTO context.outbox(event_id) VALUES($1)", event)
		if err := db.QueryRow("INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until) VALUES($1,$2,'node-1',$3,'At your practice place','place',now()+interval '1 hour') RETURNING id", user, track, event).Scan(&offer); err != nil {
			t.Fatal(err)
		}
		exec("INSERT INTO opportunity.notification_outbox(offer_id,user_id,state,provider_ticket,push_token_hash,receipt_status) VALUES($1,$2,'sent',$3,'private-hash','accepted')", offer, user, "private-ticket-"+user)
	}
	r := httptest.NewRequest("GET", "/v1/me/export", nil)
	r.Header.Set("X-User-ID", owner)
	w := httptest.NewRecorder()
	(Server{DB: db}).exportData(w, r)
	if w.Code != 200 {
		t.Fatalf("export status %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private export may be cached")
	}
	var result struct {
		SchemaVersion int                        `json:"schemaVersion"`
		Data          map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.SchemaVersion != 1 || len(result.Data) != len(exportSections) {
		t.Fatalf("incomplete export: version %d, sections %d", result.SchemaVersion, len(result.Data))
	}
	for _, section := range exportSections {
		if section.name == "profile" {
			var p struct {
				UserID string `json:"user_id"`
			}
			if err := json.Unmarshal(result.Data[section.name], &p); err != nil || p.UserID != owner {
				t.Fatalf("profile = %s (%v)", result.Data[section.name], err)
			}
			continue
		}
		var rows []json.RawMessage
		if err := json.Unmarshal(result.Data[section.name], &rows); err != nil || len(rows) != 1 {
			t.Fatalf("%s: expected one owner row, got %s (%v)", section.name, result.Data[section.name], err)
		}
		if section.name == "devices" {
			var device map[string]any
			if err := json.Unmarshal(rows[0], &device); err != nil || device["iosCalendarPermission"] != "not_requested" || device["googleCalendarPermission"] != "not_requested" {
				t.Fatalf("calendar permissions missing from device export: %s (%v)", rows[0], err)
			}
		}
	}
	for _, secret := range []string{other, "ExpoPushToken[" + owner + "]", "private-key-" + owner, "private-ticket-" + owner, "private-hash", "internal-response"} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatalf("export leaked %q", secret)
		}
	}
}

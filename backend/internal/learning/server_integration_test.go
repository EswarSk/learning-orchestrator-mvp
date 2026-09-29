package learning

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
)

func TestLearningFlow(t *testing.T) {
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
	catalog, err := LoadCatalog("../../../content")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("INTERNAL_TOKEN", "integration-test-internal-token-123456")
	var modelCalls atomic.Int32
	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/responses":
			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Error("missing OpenAI key")
			}
			var request struct {
				Instructions string `json:"instructions"`
				Input        string `json:"input"`
				Text         any    `json:"text"`
				Moderation   any    `json:"moderation"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Error(err)
			}
			text := `{"tutor":"That works. You can continue.","understandingMet":true}`
			response := map[string]any{"status": "completed", "moderation": map[string]any{"input": map[string]any{"type": "moderation_result", "flagged": false}, "output": map[string]any{"type": "moderation_result", "flagged": false}}}
			if request.Moderation == nil {
				t.Error("AI request lacked moderation")
			}
			if strings.Contains(request.Instructions, "Choose the next activity ID") {
				text = `{"nodeId":"m1-01"}`
			} else {
				modelCalls.Add(1)
				if request.Text == nil {
					t.Error("practice request lacked an assessment schema")
				}
				var input struct {
					LearnerText string `json:"learnerText"`
				}
				if err := json.Unmarshal([]byte(request.Input), &input); err != nil {
					t.Error(err)
				}
				if input.LearnerText == "I do not know" {
					text = `{"tutor":"Explain how you would do it safely.","understandingMet":false}`
				}
			}
			response["output"] = []any{map[string]any{"content": []any{map[string]any{"type": "output_text", "text": text}}}}
			_ = json.NewEncoder(w).Encode(response)
		case "/v1/text-to-speech/test-voice":
			if r.Header.Get("xi-api-key") != "test-voice-key" {
				t.Error("missing ElevenLabs key")
			}
			_, _ = w.Write([]byte("mp3"))
		case "/v1/speech-to-text":
			if r.Header.Get("xi-api-key") != "test-voice-key" {
				t.Error("missing ElevenLabs key")
			}
			_, _ = w.Write([]byte(`{"text":"spoken answer"}`))
		default:
			t.Errorf("unexpected AI request: %s", r.URL.Path)
			w.WriteHeader(400)
		}
	}))
	t.Cleanup(aiServer.Close)
	handler := Server{DB: db, Catalog: catalog, AI: AI{OpenAIKey: "test-key", OpenAIModel: "test-model", OpenAIBase: aiServer.URL, ElevenLabsKey: "test-voice-key", ElevenLabsVoiceID: "test-voice", ElevenLabsBase: aiServer.URL}}.Handler()
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	suffix := hex.EncodeToString(nonce[:])
	owner, other := "learning-test-"+suffix, "learning-test-other-"+suffix
	t.Cleanup(func() {
		for _, statement := range []string{
			"DELETE FROM learning.course_generation_quota WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.evidence WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.sessions WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.idempotency WHERE user_id IN ($1,$2)",
			"DELETE FROM learning.tracks WHERE user_id IN ($1,$2)",
		} {
			if _, err := db.Exec(statement, owner, other); err != nil {
				t.Errorf("test cleanup: %v", err)
			}
		}
	})
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM opportunity.offers WHERE user_id=$1", owner)
		_, _ = db.Exec("DELETE FROM context.events WHERE user_id=$1", owner)
		_, _ = db.Exec("DELETE FROM experience.profiles WHERE user_id=$1", owner)
	})
	call := func(user, method, path string, input any) *httptest.ResponseRecorder {
		t.Helper()
		var body []byte
		if input != nil {
			encoded, marshalErr := json.Marshal(input)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			body = encoded
		}
		r := httptest.NewRequest(method, path, bytes.NewReader(body))
		r.Header.Set("X-Internal-Token", "integration-test-internal-token-123456")
		r.Header.Set("X-User-ID", user)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	assertStatus := func(w *httptest.ResponseRecorder, want int) {
		t.Helper()
		if w.Code != want {
			t.Fatalf("status = %d, want %d: %s", w.Code, want, w.Body.String())
		}
	}

	assertStatus(call(owner, http.MethodPost, "/internal/tracks", map[string]string{"subjectId": "guitar", "goal": "Play a song"}), 422)
	created := call(owner, http.MethodPost, "/internal/tracks", map[string]string{"subjectId": "spanish", "goal": "Practice " + suffix})
	assertStatus(created, 201)
	var track Track
	if err := json.Unmarshal(created.Body.Bytes(), &track); err != nil {
		t.Fatal(err)
	}
	request := map[string]string{"trackId": track.ID, "nodeId": "m1-01", "sourceType": "curriculum", "sourceId": "m1-01", "mode": "text", "idempotencyKey": suffix}
	assertStatus(call(other, http.MethodPost, "/internal/sessions", request), 404)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions", map[string]string{"trackId": track.ID, "nodeId": "m2-01", "sourceType": "curriculum", "sourceId": "m2-01", "mode": "text", "idempotencyKey": suffix + "-locked"}), 409)
	started := call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(started, 201)
	var session struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(started.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	replayed := call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(replayed, 201)
	var repeat struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(replayed.Body.Bytes(), &repeat); err != nil {
		t.Fatal(err)
	}
	if repeat.ID != session.ID {
		t.Fatal("idempotent start returned a different session")
	}
	request["nodeId"] = "m1-02"
	assertStatus(call(owner, http.MethodPost, "/internal/sessions", request), 409)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 409)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "userEnded"}), 202)
	progress := call(owner, http.MethodGet, "/internal/progress", nil)
	assertStatus(progress, 200)
	var result struct {
		CompletedNodes  int `json:"completedNodes"`
		PracticedSkills int `json:"practicedSkills"`
		AppliedSkills   int `json:"appliedSkills"`
	}
	if err := json.Unmarshal(progress.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.CompletedNodes != 0 {
		t.Fatalf("abandoned session advanced progress: %d", result.CompletedNodes)
	}
	request["nodeId"] = "m1-01"
	request["idempotencyKey"] = suffix + "-completed"
	started = call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(started, 201)
	if err := json.Unmarshal(started.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	firstAttempt := map[string]string{"messageId": "answer-1", "text": "I do not know", "assistanceRequested": "none"}
	turn := call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", firstAttempt)
	assertStatus(turn, 200)
	if !strings.Contains(turn.Body.String(), `"understandingMet":false`) {
		t.Fatal("weak answer passed the understanding check")
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", firstAttempt), 200)
	if modelCalls.Load() != 1 {
		t.Fatalf("duplicate turn called model %d times", modelCalls.Load())
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 409)
	progress = call(owner, http.MethodGet, "/internal/progress", nil)
	assertStatus(progress, 200)
	if err := json.Unmarshal(progress.Body.Bytes(), &result); err != nil || result.CompletedNodes != 0 {
		t.Fatalf("weak answer advanced progress: %s, %v", progress.Body.String(), err)
	}
	turn = call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", map[string]string{"messageId": "answer-2", "text": "Hola, me llamo Ana.", "assistanceRequested": "model_answer"})
	assertStatus(turn, 200)
	if !strings.Contains(turn.Body.String(), `"understandingMet":false`) {
		t.Fatal("model answer incorrectly passed the understanding check")
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 409)
	turn = call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", map[string]string{"messageId": "answer-3", "text": "Hola, me llamo Ana.", "assistanceRequested": "none"})
	assertStatus(turn, 200)
	if !strings.Contains(turn.Body.String(), `"understandingMet":true`) {
		t.Fatal("correct answer did not pass the understanding check")
	}
	request["idempotencyKey"] = suffix + "-resume"
	resumed := call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(resumed, 201)
	var resume struct {
		ID               string `json:"id"`
		UnderstandingMet bool   `json:"understandingMet"`
	}
	if err := json.Unmarshal(resumed.Body.Bytes(), &resume); err != nil || resume.ID != session.ID || !resume.UnderstandingMet {
		t.Fatalf("resumed session lost assessment: %s, %v", resumed.Body.String(), err)
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 202)
	var stage string
	var reviewAt time.Time
	if err := db.QueryRow("SELECT stage,next_review_at FROM learning.skill_state WHERE track_id=$1", track.ID).Scan(&stage, &reviewAt); err != nil || stage != "needs_support" || time.Until(reviewAt) < 23*time.Hour || time.Until(reviewAt) > 25*time.Hour {
		t.Fatalf("assisted understanding should be reviewed tomorrow: %s %s %v", stage, reviewAt, err)
	}
	progress = call(owner, http.MethodGet, "/internal/progress", nil)
	assertStatus(progress, 200)
	if err := json.Unmarshal(progress.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.CompletedNodes != 1 {
		t.Fatalf("completed session did not advance progress: %d", result.CompletedNodes)
	}
	otherProgress := call(other, http.MethodGet, "/internal/progress", nil)
	assertStatus(otherProgress, 200)
	if err := json.Unmarshal(otherProgress.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.CompletedNodes != 0 {
		t.Fatalf("another user's progress changed: %d", result.CompletedNodes)
	}
	request["mode"] = "voice"
	request["idempotencyKey"] = suffix + "-voice"
	started = call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(started, 201)
	if err := json.Unmarshal(started.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	assertStatus(call(other, http.MethodPost, "/internal/sessions/"+session.ID+"/speech", map[string]string{}), 404)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/speech", map[string]string{}), 200)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/transcribe", map[string]string{"audioBase64": "YXVkaW8=", "mimeType": "audio/mp4"}), 200)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "userEnded"}), 202)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/speech", map[string]string{}), 409)

	assertStatus(call(owner, http.MethodPost, "/internal/tracks", map[string]string{"subjectName": "Guitar", "goal": "Play a song " + suffix}), 422)
	generated := call(owner, http.MethodPost, "/internal/tracks", map[string]string{"subjectId": "spanish", "goal": "Practice Spanish " + suffix})
	assertStatus(generated, 201)
	var guitar Track
	if err := json.Unmarshal(generated.Body.Bytes(), &guitar); err != nil {
		t.Fatal(err)
	}
	if guitar.SubjectTitle != "Spanish" || guitar.ReviewStatus != "unreviewed" {
		t.Fatalf("course provenance missing: %+v", guitar)
	}
	assertStatus(call(owner, http.MethodPost, "/internal/tracks", map[string]string{"subjectId": "spanish", "goal": "Practice Spanish " + suffix}), 201)
	curriculum := call(owner, http.MethodGet, "/internal/curriculum?trackId="+guitar.ID, nil)
	assertStatus(curriculum, 200)
	if !strings.Contains(curriculum.Body.String(), `"reviewStatus":"unreviewed"`) {
		t.Fatal("course disclosure missing")
	}
	request = map[string]string{"trackId": guitar.ID, "nodeId": "m1-01", "sourceType": "curriculum", "sourceId": "m1-01", "mode": "text", "idempotencyKey": suffix + "-guitar"}
	started = call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(started, 201)
	if err := json.Unmarshal(started.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", map[string]string{"messageId": "guitar-1", "text": "I do not know", "assistanceRequested": "none"}), 200)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 409)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", map[string]string{"messageId": "guitar-2", "text": "Hola, ¿cómo estás?", "assistanceRequested": "none"}), 200)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 202)
	if err := db.QueryRow("SELECT stage,next_review_at FROM learning.skill_state WHERE track_id=$1", guitar.ID).Scan(&stage, &reviewAt); err != nil || stage != "practicing" || time.Until(reviewAt) < 71*time.Hour || time.Until(reviewAt) > 73*time.Hour {
		t.Fatalf("independent understanding should get a longer review interval: %s %s %v", stage, reviewAt, err)
	}
	if _, err := db.Exec("UPDATE learning.skill_state SET next_review_at=now()-interval '1 hour' WHERE track_id=$1", guitar.ID); err != nil {
		t.Fatal(err)
	}
	moment := call(owner, http.MethodGet, "/internal/candidates?category=practice_place", nil)
	assertStatus(moment, 200)
	var choice struct {
		Candidate struct {
			TrackID    string `json:"trackId"`
			NodeID     string `json:"nodeId"`
			ReasonCode string `json:"reasonCode"`
		} `json:"candidate"`
	}
	if err := json.Unmarshal(moment.Body.Bytes(), &choice); err != nil {
		t.Fatal(err)
	}
	if choice.Candidate.TrackID != guitar.ID || choice.Candidate.NodeID != "m1-01" || choice.Candidate.ReasonCode != "due_review_at_saved_place" {
		t.Fatalf("due review should win across learning paths: %s", moment.Body.String())
	}
	calendarMoment := call(owner, http.MethodGet, "/internal/candidates?category=calendar_availability", nil)
	assertStatus(calendarMoment, 200)
	if err := json.Unmarshal(calendarMoment.Body.Bytes(), &choice); err != nil || choice.Candidate.TrackID != guitar.ID || choice.Candidate.NodeID != "m1-01" || choice.Candidate.ReasonCode != "review_due_in_calendar_gap" {
		t.Fatalf("calendar availability did not use the due review: %s, %v", calendarMoment.Body.String(), err)
	}
	otherContext := call(owner, http.MethodGet, "/internal/candidates?category=calendar_event", nil)
	assertStatus(otherContext, 200)
	if !strings.Contains(otherContext.Body.String(), `"candidate":null`) {
		t.Fatalf("an unmatched event should not invent a context match: %s", otherContext.Body.String())
	}
	selectedMoment := call(owner, http.MethodGet, "/internal/candidates?category=practice_place&trackId="+track.ID, nil)
	assertStatus(selectedMoment, 200)
	if err := json.Unmarshal(selectedMoment.Body.Bytes(), &choice); err != nil || choice.Candidate.TrackID != track.ID || choice.Candidate.NodeID != "m1-02" {
		t.Fatalf("saved place suggested a different learning path: %s, %v", selectedMoment.Body.String(), err)
	}
	if err := json.Unmarshal(moment.Body.Bytes(), &choice); err != nil {
		t.Fatal(err)
	}
	bootstrap := call(owner, http.MethodGet, "/internal/summary?trackId="+guitar.ID, nil)
	assertStatus(bootstrap, 200)
	if !strings.Contains(bootstrap.Body.String(), `"nextNode":{"id":"m1-01"`) || !strings.Contains(bootstrap.Body.String(), `"nextReason":"review_due"`) {
		t.Fatalf("Today and opportunity disagree about due review: %s", bootstrap.Body.String())
	}
	if _, err := db.Exec("INSERT INTO experience.profiles(user_id) VALUES($1)", owner); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'location_context','1',true,'grant')", owner); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'calendar_context','1',true,'grant-calendar')", owner); err != nil {
		t.Fatal(err)
	}
	makeOffer := func() string {
		t.Helper()
		var eventID, offerID string
		if err := db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,'ios_geofence',gen_random_uuid()::text,'test-place','practice_place',now(),now()+interval '30 minutes') RETURNING id`, owner).Scan(&eventID); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow(`INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until)
 VALUES($1,$2,$3,$4,'Apply this skill','saved_practice_place',now()+interval '30 minutes') RETURNING id`, owner, choice.Candidate.TrackID, choice.Candidate.NodeID, eventID).Scan(&offerID); err != nil {
			t.Fatal(err)
		}
		return offerID
	}
	originalOffer := makeOffer()
	request = map[string]string{"trackId": guitar.ID, "nodeId": "m1-01", "sourceType": "opportunity", "sourceId": originalOffer, "mode": "text", "idempotencyKey": suffix + "-opportunity"}
	started = call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(started, 201)
	if err := json.Unmarshal(started.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "userEnded"}), 202)
	application := map[string]string{"opportunityId": originalOffer, "trackId": choice.Candidate.TrackID, "nodeId": choice.Candidate.NodeID, "outcome": "applied"}
	withForgedCategory := map[string]string{"opportunityId": originalOffer, "trackId": choice.Candidate.TrackID, "nodeId": choice.Candidate.NodeID, "outcome": "applied", "contextCategory": "practice_place"}
	assertStatus(call(owner, http.MethodPost, "/internal/applications", withForgedCategory), 422)
	assertStatus(call(owner, http.MethodPost, "/internal/applications", application), 201)
	assertStatus(call(owner, http.MethodPost, "/internal/applications", application), 201)
	var calendarEventID, calendarOfferID string
	if err := db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,'calendar',gen_random_uuid()::text,'test-calendar','calendar_availability',now(),now()+interval '30 minutes') RETURNING id`, owner).Scan(&calendarEventID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until)
 VALUES($1,$2,$3,$4,'Try this review','review_due_in_calendar_gap',now()+interval '30 minutes') RETURNING id`, owner, guitar.ID, "m1-01", calendarEventID).Scan(&calendarOfferID); err != nil {
		t.Fatal(err)
	}
	calendarReport := map[string]string{"opportunityId": calendarOfferID, "trackId": guitar.ID, "nodeId": "m1-01", "outcome": "missed"}
	assertStatus(call(owner, http.MethodPost, "/internal/applications", calendarReport), 201)
	if _, err := db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'ios_calendar_context','1',true,'grant-ios-calendar')", owner); err != nil {
		t.Fatal(err)
	}
	makeIOSCalendarOffer := func() string {
		t.Helper()
		var eventID, offerID string
		if err := db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,'ios_calendar',gen_random_uuid()::text,'ios:test-device','calendar_availability',now(),now()+interval '30 minutes') RETURNING id`, owner).Scan(&eventID); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow(`INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until)
 VALUES($1,$2,'m1-01',$3,'Try this review','review_due_in_calendar_gap',now()+interval '3 minutes') RETURNING id`, owner, guitar.ID, eventID).Scan(&offerID); err != nil {
			t.Fatal(err)
		}
		return offerID
	}
	iosReport := map[string]string{"opportunityId": makeIOSCalendarOffer(), "trackId": guitar.ID, "nodeId": "m1-01", "outcome": "missed"}
	assertStatus(call(owner, http.MethodPost, "/internal/applications", iosReport), 201)
	if _, err := db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'ios_calendar_context','1',false,'revoke-ios-calendar')", owner); err != nil {
		t.Fatal(err)
	}
	iosReport["opportunityId"] = makeIOSCalendarOffer()
	assertStatus(call(owner, http.MethodPost, "/internal/applications", iosReport), 409)
	blockedOffer := makeOffer()
	if _, err := db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'location_context','1',false,'revoke')", owner); err != nil {
		t.Fatal(err)
	}
	application["opportunityId"] = blockedOffer
	assertStatus(call(owner, http.MethodPost, "/internal/applications", application), 409)
	request = map[string]string{"trackId": guitar.ID, "nodeId": "m1-01", "sourceType": "opportunity", "sourceId": blockedOffer, "mode": "text", "idempotencyKey": suffix + "-revoked"}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions", request), 409)
	application["opportunityId"] = originalOffer
	assertStatus(call(owner, http.MethodPost, "/internal/applications", application), 201)
	request = map[string]string{"trackId": guitar.ID, "nodeId": "m1-01", "sourceType": "curriculum", "sourceId": "m1-01", "mode": "text"}
	if err := db.QueryRow("SELECT stage,next_review_at FROM learning.skill_state WHERE track_id=$1", guitar.ID).Scan(&stage, &reviewAt); err != nil || stage != "practicing" || reviewAt.After(time.Now()) {
		t.Fatalf("self-report postponed assessed review: %s %s %v", stage, reviewAt, err)
	}
	request["idempotencyKey"] = suffix + "-review"
	started = call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(started, 201)
	if err := json.Unmarshal(started.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", map[string]string{"messageId": "guitar-review", "text": "I would play the chord safely.", "assistanceRequested": "none"}), 200)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 202)
	moment = call(owner, http.MethodGet, "/internal/candidates?category=practice_place", nil)
	assertStatus(moment, 200)
	if err := json.Unmarshal(moment.Body.Bytes(), &choice); err != nil || choice.Candidate.TrackID != track.ID || choice.Candidate.NodeID != "m1-02" || choice.Candidate.ReasonCode != "saved_practice_place" {
		t.Fatalf("successful review did not change next opportunity: %s, %v", moment.Body.String(), err)
	}
	calendarMoment = call(owner, http.MethodGet, "/internal/candidates?category=calendar_availability", nil)
	assertStatus(calendarMoment, 200)
	if err := json.Unmarshal(calendarMoment.Body.Bytes(), &choice); err != nil || choice.Candidate.TrackID != track.ID || choice.Candidate.NodeID != "m1-02" || choice.Candidate.ReasonCode != "calendar_availability" {
		t.Fatalf("calendar availability did not adapt after review: %s, %v", calendarMoment.Body.String(), err)
	}
	if _, err := db.Exec("INSERT INTO experience.consents(user_id,purpose,policy_version,granted,request_key) VALUES($1,'location_context','1',true,'restore-after-review')", owner); err != nil {
		t.Fatal(err)
	}
	missedOffer := makeOffer()
	missed := map[string]string{"opportunityId": missedOffer, "trackId": track.ID, "nodeId": "m1-02", "outcome": "missed"}
	assertStatus(call(owner, http.MethodPost, "/internal/applications", missed), 201)
	progress = call(owner, http.MethodGet, "/internal/progress", nil)
	assertStatus(progress, 200)
	if err := json.Unmarshal(progress.Body.Bytes(), &result); err != nil || result.PracticedSkills != 2 || result.AppliedSkills != 1 {
		t.Fatalf("missed opportunity counted as practice or application: %s, %v", progress.Body.String(), err)
	}
	moment = call(owner, http.MethodGet, "/internal/candidates?category=practice_place&trackId="+track.ID, nil)
	assertStatus(moment, 200)
	if err := json.Unmarshal(moment.Body.Bytes(), &choice); err != nil || choice.Candidate.NodeID != "m1-02" {
		t.Fatalf("missed opportunity advanced the curriculum: %s, %v", moment.Body.String(), err)
	}
	practiceOffer := makeOffer()
	request = map[string]string{"trackId": track.ID, "nodeId": "m1-02", "sourceType": "opportunity", "sourceId": practiceOffer, "mode": "text", "idempotencyKey": suffix + "-opportunity-complete"}
	started = call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(started, 201)
	if err := json.Unmarshal(started.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", map[string]string{"messageId": "spanish-opportunity", "text": "I can ask a follow-up question.", "assistanceRequested": "none"}), 200)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 202)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 202)
	progress = call(owner, http.MethodGet, "/internal/progress", nil)
	assertStatus(progress, 200)
	if err := json.Unmarshal(progress.Body.Bytes(), &result); err != nil || result.CompletedNodes != 3 {
		t.Fatalf("assessed opportunity did not advance the course exactly once: %s, %v", progress.Body.String(), err)
	}
	moment = call(owner, http.MethodGet, "/internal/candidates?category=practice_place&trackId="+track.ID, nil)
	assertStatus(moment, 200)
	if err := json.Unmarshal(moment.Body.Bytes(), &choice); err != nil || choice.Candidate.NodeID != "m1-03" {
		t.Fatalf("next opportunity repeated a completed activity: %s, %v", moment.Body.String(), err)
	}
	templateID := "test-application-" + suffix
	if _, err := db.Exec(`INSERT INTO learning.application_templates(id,subject_id,skill_id,context_category,title,prompt,reviewed)
 VALUES($1,'spanish','follow-up','calendar_event','Ask a follow-up','Ask a useful follow-up question.',true)`, templateID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM learning.application_templates WHERE id=$1", templateID) })
	contextMoment := call(owner, http.MethodGet, "/internal/candidates?category=calendar_event&trackId="+track.ID, nil)
	assertStatus(contextMoment, 200)
	if err := json.Unmarshal(contextMoment.Body.Bytes(), &choice); err != nil || choice.Candidate.NodeID != templateID || choice.Candidate.ReasonCode != "approved_context_match" {
		t.Fatalf("reviewed contextual activity was not selected: %s, %v", contextMoment.Body.String(), err)
	}
	var templateEventID, templateOfferID string
	if err := db.QueryRow(`INSERT INTO context.events(user_id,source,source_event_id,source_resource_id,category,observed_at,valid_until)
 VALUES($1,'calendar',gen_random_uuid()::text,'test-calendar','calendar_event',now(),now()+interval '30 minutes') RETURNING id`, owner).Scan(&templateEventID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO opportunity.offers(user_id,track_id,node_id,context_event_id,context_label,reason_code,valid_until)
 VALUES($1,$2,$3,$4,'Ask a follow-up','approved_context_match',now()+interval '30 minutes') RETURNING id`, owner, track.ID, templateID, templateEventID).Scan(&templateOfferID); err != nil {
		t.Fatal(err)
	}
	request = map[string]string{"trackId": track.ID, "nodeId": templateID, "sourceType": "opportunity", "sourceId": templateOfferID, "mode": "text", "idempotencyKey": suffix + "-template-practice"}
	started = call(owner, http.MethodPost, "/internal/sessions", request)
	assertStatus(started, 201)
	if err := json.Unmarshal(started.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/turns", map[string]string{"messageId": "template-practice", "text": "Can you tell me more?", "assistanceRequested": "none"}), 200)
	assertStatus(call(owner, http.MethodPost, "/internal/sessions/"+session.ID+"/finish", map[string]string{"reason": "completed"}), 202)
	progress = call(owner, http.MethodGet, "/internal/progress", nil)
	assertStatus(progress, 200)
	if err := json.Unmarshal(progress.Body.Bytes(), &result); err != nil || result.CompletedNodes != 3 {
		t.Fatalf("context-only activity incorrectly advanced the course: %s, %v", progress.Body.String(), err)
	}
	assertStatus(call(owner, http.MethodPost, "/internal/tracks", map[string]string{"subjectName": "Guitar", "goal": "One more riff " + suffix}), 422)
}

func TestConcurrentStartsRespectFreeCap(t *testing.T) {
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
	owner := "free-cap-test-" + hex.EncodeToString(nonce[:])
	defer func() {
		_, _ = db.Exec("DELETE FROM learning.idempotency WHERE user_id=$1", owner)
		_, _ = db.Exec("DELETE FROM learning.sessions WHERE user_id=$1", owner)
		_, _ = db.Exec("DELETE FROM learning.tracks WHERE user_id=$1", owner)
	}()
	var track string
	if err = db.QueryRow("INSERT INTO learning.tracks(user_id,subject_id,goal) VALUES($1,'spanish','Test cap') RETURNING id", owner).Scan(&track); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO learning.sessions(user_id,track_id,node_id,source_type,source_id,mode,state)
 SELECT $1,$2,'m1-01','curriculum','existing-'||n,'text','completed' FROM generate_series(1,24) n`, owner, track); err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog("../../../content")
	if err != nil {
		t.Fatal(err)
	}
	s := Server{DB: db, Catalog: catalog, AI: AI{OpenAIKey: "test", OpenAIModel: "test"}}
	start := make(chan struct{})
	results := make(chan int, 30)
	var workers sync.WaitGroup
	for i := 0; i < 30; i++ {
		workers.Add(1)
		go func(i int) {
			defer workers.Done()
			<-start
			body := fmt.Sprintf(`{"trackId":%q,"nodeId":"m1-01","sourceType":"curriculum","sourceId":%q,"mode":"text","idempotencyKey":%q}`, track, fmt.Sprintf("attempt-%d", i), fmt.Sprintf("request-%d", i))
			r := httptest.NewRequest(http.MethodPost, "/internal/sessions", strings.NewReader(body))
			r.Header.Set("X-User-ID", owner)
			w := httptest.NewRecorder()
			s.startSession(w, r)
			results <- w.Code
		}(i)
	}
	close(start)
	workers.Wait()
	close(results)
	accepted, capped := 0, 0
	for status := range results {
		switch status {
		case http.StatusCreated:
			accepted++
		case http.StatusPaymentRequired:
			capped++
		default:
			t.Fatalf("unexpected concurrent start status %d", status)
		}
	}
	var used int
	if err = db.QueryRow("SELECT count(*) FROM learning.sessions WHERE user_id=$1 AND state IN ('active','completed')", owner).Scan(&used); err != nil || accepted != 1 || capped != 29 || used != 25 {
		t.Fatalf("cap exceeded: accepted=%d capped=%d used=%d query=%v", accepted, capped, used, err)
	}
}

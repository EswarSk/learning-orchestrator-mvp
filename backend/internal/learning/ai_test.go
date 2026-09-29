package learning

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIReplyUsesBoundedPrivateResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("unexpected request: %s", r.URL.Path)
		}
		var body struct {
			Model string `json:"model"`
			Store bool   `json:"store"`
			Input string `json:"input"`
			Text  struct {
				Format struct {
					Type string `json:"type"`
					Name string `json:"name"`
				} `json:"format"`
			} `json:"text"`
			Moderation struct {
				Model string `json:"model"`
			} `json:"moderation"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Model != "test-model" || body.Store || !strings.Contains(body.Input, "guitar") || body.Text.Format.Type != "json_schema" || body.Text.Format.Name != "practice_check" || body.Moderation.Model != "omni-moderation-latest" {
			t.Errorf("wrong request: %+v", body)
		}
		_, _ = io.WriteString(w, `{"status":"completed","moderation":{"input":{"type":"moderation_result","flagged":false},"output":{"type":"moderation_result","flagged":false}},"output":[{"type":"reasoning"},{"type":"message","content":[{"type":"output_text","text":"{\"tutor\":\"Try a slower strum.\",\"understandingMet\":false}"}]}]}`)
	}))
	defer server.Close()
	a := AI{OpenAIKey: "secret", OpenAIModel: "test-model", OpenAIBase: server.URL}
	reply, err := a.reply(context.Background(), "Guitar", Node{Prompt: "Play a chord"}, nil, "guitar", "none")
	if err != nil || reply.Tutor != "Try a slower strum." || reply.UnderstandingMet {
		t.Fatalf("reply=%+v err=%v", reply, err)
	}
}

func TestOpenAIReplyFailsClosedOnModeration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"completed","moderation":{"input":{"type":"moderation_result","flagged":false},"output":{"type":"moderation_result","flagged":true}},"output":[{"content":[{"type":"output_text","text":"{\"tutor\":\"Unsafe response\",\"understandingMet\":true}"}]}]}`)
	}))
	defer server.Close()
	a := AI{OpenAIKey: "secret", OpenAIModel: "test-model", OpenAIBase: server.URL}
	if _, err := a.reply(context.Background(), "Guitar", Node{Prompt: "Play a chord"}, nil, "answer", "none"); err == nil {
		t.Fatal("flagged tutor response was accepted")
	}
}

func TestPlannerRejectsUnlistedActivity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"completed","moderation":{"input":{"type":"moderation_result","flagged":false},"output":{"type":"moderation_result","flagged":false}},"output":[{"content":[{"type":"output_text","text":"{\"nodeId\":\"invented\"}"}]}]}`)
	}))
	defer server.Close()
	a := AI{OpenAIKey: "secret", OpenAIModel: "test-model", OpenAIBase: server.URL}
	_, err := a.chooseActivity(context.Background(), "learn", []activityChoice{{Node: Node{ID: "fixed"}}}, map[string]bool{}, nil)
	if err == nil {
		t.Fatal("planner accepted an unlisted activity")
	}
}

func TestElevenLabsVoiceContracts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("xi-api-key") != "secret" {
			t.Error("missing server-side key")
		}
		switch r.URL.Path {
		case "/v1/text-to-speech/test-voice":
			var body struct {
				Text    string `json:"text"`
				ModelID string `json:"model_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body.Text != "Hello" || body.ModelID != "eleven_multilingual_v2" {
				t.Errorf("wrong speech request: %+v", body)
			}
			_, _ = io.WriteString(w, "mp3-data")
		case "/v1/speech-to-text":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Error(err)
			}
			if r.FormValue("model_id") != "scribe_v2" {
				t.Error("wrong transcript model")
			}
			file, _, err := r.FormFile("file")
			if err != nil {
				t.Error(err)
			} else {
				file.Close()
			}
			_, _ = io.WriteString(w, `{"text":"Hello there"}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	a := AI{ElevenLabsKey: "secret", ElevenLabsVoiceID: "test-voice", ElevenLabsBase: server.URL}
	audio, err := a.speech(context.Background(), "Hello")
	if err != nil || string(audio) != "mp3-data" {
		t.Fatalf("audio=%q err=%v", audio, err)
	}
	transcript, err := a.transcribe(context.Background(), "practice.m4a", []byte("audio-data"))
	if err != nil || transcript != "Hello there" {
		t.Fatalf("transcript=%q err=%v", transcript, err)
	}
}

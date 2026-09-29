package learning

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestSubjectsSorted(t *testing.T) {
	s := Server{Catalog: Catalog{
		"spanish": {Subject: Subject{ID: "spanish", Title: "Spanish"}},
		"guitar":  {Subject: Subject{ID: "guitar", Title: "Guitar"}},
	}}
	w := httptest.NewRecorder()
	s.subjects(w, httptest.NewRequest("GET", "/internal/subjects", nil))
	var body struct {
		Items []Subject `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || len(body.Items) != 2 || body.Items[0].ID != "guitar" || body.Items[1].ID != "spanish" {
		t.Fatalf("subjects were not sorted: %s", w.Body.String())
	}
}

func TestPracticeRequiresConfiguredProvider(t *testing.T) {
	s := Server{}
	body := []byte(`{"nodeId":"m1-01","sourceId":"m1-01","sourceType":"curriculum","mode":"text","idempotencyKey":"request-1"}`)
	w := httptest.NewRecorder()
	s.startSession(w, httptest.NewRequest("POST", "/internal/sessions", bytes.NewReader(body)))
	if w.Code != 503 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

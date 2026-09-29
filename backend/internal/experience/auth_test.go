package experience

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"learning-orchestrator/backend/internal/platform"
)

func TestLocalAuthRejectsSpoofing(t *testing.T) {
	handler := Verifier{Local: true}.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := platform.User(r); got != "00000000-0000-4000-8000-000000000001" {
			t.Errorf("trusted user = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, tc := range []struct {
		name, token string
		want        int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"wrong token", "Bearer wrong", http.StatusUnauthorized},
		{"local token", "Bearer local-development", http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/bootstrap", nil)
			r.Header.Set("X-User-ID", "another-user")
			if tc.token != "" {
				r.Header.Set("Authorization", tc.token)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d", w.Code, tc.want)
			}
		})
	}
}

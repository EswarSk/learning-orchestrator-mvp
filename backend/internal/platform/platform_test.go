package platform

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRejectsTrailingAndOversize(t *testing.T) {
	for _, body := range []string{`{"a":1} {"a":2}`, `{"a":1} nope`, strings.Repeat(" ", 64*1024+1)} {
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		var value struct {
			A int `json:"a"`
		}
		if err := Decode(req, &value); err == nil {
			t.Fatalf("accepted invalid body of length %d", len(body))
		}
	}
}

func TestDevelopmentInternalTokenRejectedOutsideLocalMode(t *testing.T) {
	t.Setenv("INTERNAL_TOKEN", "local-internal-token-change-in-production")
	t.Setenv("APP_ENV", "production")
	defer func() {
		if recover() == nil {
			t.Fatal("production accepted the shared development service token")
		}
	}()
	Internal(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
}

func TestDevelopmentInternalTokenAllowedLocally(t *testing.T) {
	t.Setenv("INTERNAL_TOKEN", "local-internal-token-change-in-production")
	t.Setenv("APP_ENV", "local")
	Internal(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
}

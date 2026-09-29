package experience

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"learning-orchestrator/backend/internal/platform"
)

type authKey struct{}
type Verifier struct {
	JWT   *oidc.IDTokenVerifier
	Local bool
}

func NewVerifier(ctx context.Context) (Verifier, error) {
	if os.Getenv("APP_ENV") == "local" {
		return Verifier{Local: true}, nil
	}
	issuer, audience := os.Getenv("OIDC_ISSUER"), os.Getenv("OIDC_AUDIENCE")
	if issuer == "" || audience == "" {
		return Verifier{}, errors.New("OIDC_ISSUER and OIDC_AUDIENCE are required outside local mode")
	}
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return Verifier{}, err
	}
	return Verifier{JWT: provider.Verifier(&oidc.Config{ClientID: audience})}, nil
}

func (v Verifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			platform.Error(w, 401, "unauthorized", "Authentication required")
			return
		}
		token := strings.TrimPrefix(authorization, "Bearer ")
		user := ""
		if v.Local {
			if platform.EqualToken(token, "local-development") {
				user = "00000000-0000-4000-8000-000000000001"
			}
		} else {
			verified, err := v.JWT.Verify(r.Context(), token)
			if err == nil {
				user = verified.Subject
			}
		}
		if user == "" {
			platform.Error(w, 401, "unauthorized", "Invalid or expired authentication")
			return
		}
		// Never trust caller-supplied identity headers.
		r.Header.Set("X-User-ID", user)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authKey{}, user)))
	})
}

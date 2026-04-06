package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/chatchitganggang/internal-comm-backend/internal/user"
)

// BearerMiddleware validates Authorization: Bearer <JWT>, upserts the user row, and attaches principal to context.
func BearerMiddleware(v *Validator, repo user.Sync) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			claims, err := v.ParseAndValidate(r.Context(), raw)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			display := claims.PreferredUsername
			if display == "" {
				display = claims.Email
			}
			uid, err := repo.UpsertByKeycloakSub(r.Context(), claims.Subject, display)
			if err != nil {
				writeAuthError(w, http.StatusInternalServerError, "user sync failed")
				return
			}
			ctx := WithPrincipal(r.Context(), uid, claims.Subject, claims.Email, claims.PreferredUsername)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(h string) (string, bool) {
	const p = "Bearer "
	if len(h) < len(p) || !strings.EqualFold(h[:len(p)], p) {
		return "", false
	}
	t := strings.TrimSpace(h[len(p):])
	if t == "" {
		return "", false
	}
	return t, true
}

func writeAuthError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

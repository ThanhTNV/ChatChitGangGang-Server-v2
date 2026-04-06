package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
)

func handleMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	sub, _ := auth.Subject(r.Context())
	email, _ := auth.Email(r.Context())
	pref, _ := auth.PreferredUsername(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id":             uid.String(),
		"sub":                 sub,
		"email":               email,
		"preferred_username": pref,
	})
}

package httpserver

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/user"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // tighten for production (Origin allowlist)
	},
}

// HandleWebSocket validates the token (query or Sec-WebSocket-Protocol), upserts user, upgrades, stub read loop.
func HandleWebSocket(log *slog.Logger, v *auth.Validator, repo user.Sync) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := wsAccessToken(r)
		if raw == "" {
			http.Error(w, "missing access token", http.StatusUnauthorized)
			return
		}
		claims, err := v.ParseAndValidate(r.Context(), raw)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		display := claims.PreferredUsername
		if display == "" {
			display = claims.Email
		}
		if _, err := repo.UpsertByKeycloakSub(r.Context(), claims.Subject, display); err != nil {
			http.Error(w, "user sync failed", http.StatusInternalServerError)
			return
		}

		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Debug("ws upgrade", "error", err)
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				log.Debug("ws close", "error", err)
			}
		}()

		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
	}
}

func wsAccessToken(r *http.Request) string {
	if t := r.URL.Query().Get("access_token"); t != "" {
		return strings.TrimSpace(t)
	}
	for _, header := range r.Header.Values("Sec-WebSocket-Protocol") {
		for _, part := range strings.Split(header, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if strings.HasPrefix(strings.ToLower(part), "bearer.") {
				return strings.TrimSpace(part[len("bearer."):])
			}
		}
	}
	return ""
}

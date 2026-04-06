package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/channel"
)

func listChannelsHandler(repo channel.Lister, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if repo == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "service misconfigured"})
			return
		}
		uid, ok := auth.UserID(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		list, err := repo.ListForUser(r.Context(), uid)
		if err != nil {
			log.LogAttrs(r.Context(), slog.LevelWarn, "list_channels_failed",
				slog.Any("error", err),
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"channels": list})
	}
}

package httpserver

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/chat"
)

func listChannelMessagesHandler(store chat.Store, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "service misconfigured"})
			return
		}
		uid, ok := auth.UserID(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		rawID := chi.URLParam(r, "channelID")
		channelID, err := uuid.Parse(rawID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid channel id"})
			return
		}

		limit := 0
		if q := r.URL.Query().Get("limit"); q != "" {
			n, err := strconv.Atoi(q)
			if err != nil || n < 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid limit"})
				return
			}
			limit = n
		}
		cursor := r.URL.Query().Get("cursor")

		res, err := store.ListForMember(r.Context(), uid, channelID, chat.ListMessagesOpts{
			Cursor: cursor,
			Limit:  limit,
		})
		if err != nil {
			switch {
			case errors.Is(err, chat.ErrNotMember):
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
			case errors.Is(err, chat.ErrInvalidCursor):
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cursor"})
			default:
				log.LogAttrs(r.Context(), slog.LevelWarn, "list_messages_failed",
					slog.Any("error", err),
				)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			}
			return
		}
		out := map[string]any{
			"messages": res.Messages,
		}
		if res.NextCursor != "" {
			out["next_cursor"] = res.NextCursor
		}
		writeJSON(w, http.StatusOK, out)
	}
}

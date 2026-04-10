package httpserver

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/channel"
)

const maxChannelNameRunes = 256

// small JSON payloads only
const maxCreateChannelBodyBytes = 65536

func listChannelsHandler(store channel.Store, log *slog.Logger) http.HandlerFunc {
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
		list, err := store.ListForUser(r.Context(), uid)
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

func createChannelHandler(store channel.Store, log *slog.Logger) http.HandlerFunc {
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
		r.Body = http.MaxBytesReader(w, r.Body, maxCreateChannelBodyBytes)
		var body channel.CreateChannelRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			if err == io.EOF {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty body"})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		typ := strings.TrimSpace(body.Type)
		if typ == "" {
			typ = "group"
		}
		if typ != "group" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": `only type "group" is supported`})
			return
		}
		name := strings.TrimSpace(body.Name)
		if name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
			return
		}
		if utf8.RuneCountInString(name) > maxChannelNameRunes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name too long"})
			return
		}
		ch, err := store.CreateGroup(r.Context(), uid, name)
		if err != nil {
			log.LogAttrs(r.Context(), slog.LevelWarn, "create_channel_failed",
				slog.Any("error", err),
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		writeJSON(w, http.StatusCreated, ch)
	}
}

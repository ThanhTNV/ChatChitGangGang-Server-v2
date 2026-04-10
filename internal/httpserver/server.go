package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chatchitganggang/internal-comm-backend/internal/apiembed"
	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/channel"
	"github.com/chatchitganggang/internal-comm-backend/internal/chat"
	"github.com/chatchitganggang/internal-comm-backend/internal/config"
	"github.com/chatchitganggang/internal-comm-backend/internal/user"
)

// Auth wires optional Keycloak/JWT routes. When Bearer is nil, /v1 is not mounted.
type Auth struct {
	Bearer    func(http.Handler) http.Handler
	Validator *auth.Validator
	Users     user.Sync
	Channels  channel.Store
	Messages  chat.Store
}

// New builds an HTTP server with routing and production-oriented timeouts.
// db may be nil; /ready marks database as skipped. Optional deps (e.g. Redis) extend /ready.
func New(cfg *config.Config, log *slog.Logger, db *pgxpool.Pool, a *Auth, deps Deps) *http.Server {
	r := chi.NewRouter()
	r.Use(chimw.RequestID, chimw.RealIP, requestLogger(log), chimw.Recoverer)
	r.Get("/health", health)
	r.Get("/ready", readyHandler(cfg, db, deps, log))
	RegisterOpenAPISpec(r, cfg, apiembed.OpenAPIYAML)

	if a != nil && a.Bearer != nil && a.Validator != nil && a.Users != nil {
		r.Route("/v1", func(r chi.Router) {
			// WebSocket clients typically send the token via query or Sec-WebSocket-Protocol, not Authorization.
			r.Group(func(r chi.Router) {
				r.Use(a.Bearer)
				r.Get("/me", handleMe)
				r.Get("/channels", listChannelsHandler(a.Channels, log))
				r.Post("/channels", createChannelHandler(a.Channels, log))
				r.Get("/channels/{channelID}/messages", listChannelMessagesHandler(a.Messages, log))
			})
			r.Get("/ws", HandleWebSocket(log, a.Validator, a.Users))
		})
	}

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func requestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				log.LogAttrs(r.Context(), slog.LevelInfo, "request",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.Int64("duration_ms", time.Since(start).Milliseconds()),
					slog.String("request_id", chimw.GetReqID(r.Context())),
				)
			}()
			next.ServeHTTP(ww, r)
		})
	}
}

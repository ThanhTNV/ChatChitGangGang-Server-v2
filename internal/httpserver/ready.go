package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chatchitganggang/internal-comm-backend/internal/config"
)

const readyCheckTimeout = 5 * time.Second

func readyHandler(cfg *config.Config, db *pgxpool.Pool, deps Deps, log *slog.Logger) http.HandlerFunc {
	hc := &http.Client{Timeout: readyCheckTimeout}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyCheckTimeout)
		defer cancel()

		checks := make(map[string]string)

		if db == nil {
			checks["database"] = "skipped"
		} else if err := db.Ping(ctx); err != nil {
			log.LogAttrs(r.Context(), slog.LevelWarn, "ready_check_failed",
				slog.String("check", "database"),
				slog.Any("error", err),
			)
			checks["database"] = "error"
		} else {
			checks["database"] = "ok"
		}

		if deps.Redis == nil {
			checks["redis"] = "skipped"
		} else if err := deps.Redis.Ping(ctx).Err(); err != nil {
			log.LogAttrs(r.Context(), slog.LevelWarn, "ready_check_failed",
				slog.String("check", "redis"),
				slog.Any("error", err),
			)
			checks["redis"] = "error"
		} else {
			checks["redis"] = "ok"
		}

		kcURL := keycloakProbeURL(cfg)
		if kcURL == "" {
			checks["keycloak"] = "skipped"
		} else if ok := httpProbe(ctx, hc, kcURL); !ok {
			log.LogAttrs(r.Context(), slog.LevelWarn, "ready_check_failed",
				slog.String("check", "keycloak"),
				slog.String("url", kcURL),
			)
			checks["keycloak"] = "error"
		} else {
			checks["keycloak"] = "ok"
		}

		minioURL := minioHealthURL(cfg)
		if minioURL == "" {
			checks["minio"] = "skipped"
		} else if ok := httpProbe(ctx, hc, minioURL); !ok {
			log.LogAttrs(r.Context(), slog.LevelWarn, "ready_check_failed",
				slog.String("check", "minio"),
				slog.String("url", minioURL),
			)
			checks["minio"] = "error"
		} else {
			checks["minio"] = "ok"
		}

		status := http.StatusOK
		bodyStatus := "ok"
		if anyCheckError(checks) {
			status = http.StatusServiceUnavailable
			bodyStatus = "error"
		}
		writeJSON(w, status, map[string]any{
			"status": bodyStatus,
			"checks": checks,
		})
	}
}

// keycloakProbeURL returns the URL to GET for Keycloak readiness.
// Prefer KEYCLOAK_READY_URL so containers can reach Keycloak (Keycloak 26: /health/* on management port 9000 by default).
// while KEYCLOAK_ISSUER stays the public URL that appears in JWTs (e.g. http://localhost:8090/realms/...).
func keycloakProbeURL(cfg *config.Config) string {
	if cfg.KeycloakReadyURL != "" {
		return cfg.KeycloakReadyURL
	}
	return cfg.KeycloakIssuer
}

func anyCheckError(checks map[string]string) bool {
	for _, v := range checks {
		if v == "error" {
			return true
		}
	}
	return false
}

func httpProbe(ctx context.Context, client *http.Client, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func minioHealthURL(cfg *config.Config) string {
	if cfg.MinIOEndpoint == "" {
		return ""
	}
	scheme := "http"
	if cfg.MinIOUseSSL {
		scheme = "https"
	}
	host := strings.TrimSpace(cfg.MinIOEndpoint)
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	if host == "" {
		return ""
	}
	return scheme + "://" + host + "/minio/health/live"
}

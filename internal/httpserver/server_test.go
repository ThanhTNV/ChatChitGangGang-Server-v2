package httpserver

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chatchitganggang/internal-comm-backend/internal/config"
)

func TestHealthAndReady(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{HTTPAddr: ":0", LogLevel: slog.LevelError, LogJSON: false}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(cfg, log, nil, nil, Deps{})

	for _, path := range []string{"/health", "/docs"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, rec.Code)
		}
	}

	reqReady := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recReady := httptest.NewRecorder()
	srv.Handler.ServeHTTP(recReady, reqReady)
	if recReady.Code != http.StatusOK {
		t.Fatalf("/ready: status %d", recReady.Code)
	}
	var readyBody struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.NewDecoder(recReady.Body).Decode(&readyBody); err != nil {
		t.Fatal(err)
	}
	if readyBody.Status != "ok" {
		t.Fatalf("ready status: %q", readyBody.Status)
	}
	for _, k := range []string{"database", "redis", "keycloak", "minio"} {
		if readyBody.Checks[k] != "skipped" {
			t.Fatalf("check %s: want skipped, got %q", k, readyBody.Checks[k])
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/openapi.yaml: status %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/yaml; charset=utf-8" {
		t.Fatalf("openapi Content-Type: got %q", got)
	}
	body, _ := io.ReadAll(rec.Body)
	if len(body) < 100 {
		t.Fatalf("openapi body too short: %d bytes", len(body))
	}
	if !strings.HasPrefix(string(body), "openapi: 3.0.3") {
		t.Fatalf("openapi body prefix: %q", string(body[:80]))
	}
}

package httpserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/chatchitganggang/internal-comm-backend/internal/config"
)

func TestOpenAPIPublicBase_envOverride(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{OpenAPIPublicBase: "https://example.test/api"}
	req := httptest.NewRequest("GET", "/openapi.yaml", nil)
	if got := openAPIPublicBase(cfg, req); got != "https://example.test/api" {
		t.Fatalf("got %q", got)
	}
}

func TestOpenAPIPublicBase_forwarded(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/openapi.yaml", nil)
	req.Host = "localhost:8080"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "localhost")
	req.Header.Set("X-Forwarded-Prefix", "/api")
	if got := openAPIPublicBase(nil, req); got != "https://localhost/api" {
		t.Fatalf("got %q", got)
	}
}

func TestOpenAPIPublicBase_directHost(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/openapi.yaml", nil)
	req.Host = "127.0.0.1:9090"
	if got := openAPIPublicBase(nil, req); got != "http://127.0.0.1:9090" {
		t.Fatalf("got %q", got)
	}
}

func TestOpenAPIPublicBase_emptyHost(t *testing.T) {
	t.Parallel()
	// httptest.NewRequest defaults Host to example.com; empty Host means no usable public URL.
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/openapi.yaml"},
		Header: http.Header{},
	}
	if got := openAPIPublicBase(nil, req); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestPatchOpenAPIServers(t *testing.T) {
	t.Parallel()
	spec := []byte("openapi: 3.0.3\n" + openAPIServersRelative + "paths: {}\n")
	out := patchOpenAPIServers(spec, "https://localhost/api")
	if strings.Contains(string(out), "url: /\n") {
		t.Fatalf("still relative: %s", out)
	}
	if !strings.Contains(string(out), `"https://localhost/api"`) {
		t.Fatalf("missing base: %s", out)
	}
}

func TestPatchOpenAPIServers_emptyBaseNoop(t *testing.T) {
	t.Parallel()
	spec := []byte("openapi: 3.0.3\n" + openAPIServersRelative)
	out := patchOpenAPIServers(spec, "")
	if string(out) != string(spec) {
		t.Fatal("expected unchanged")
	}
}

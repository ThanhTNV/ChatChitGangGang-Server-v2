package httpserver

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/chatchitganggang/internal-comm-backend/internal/config"
)

// openAPIServersRelative is the embedded servers block in internal/apiembed/openapi.yaml (must stay in sync).
const openAPIServersRelative = "servers:\n  - url: /\n    description: Relative to the server base URL\n"

// RegisterOpenAPISpec registers GET /openapi.yaml and GET /docs (Swagger UI).
// OpenAPI "servers" and Swagger spec URL adapt to the caller: direct (:8080), or behind a proxy
// (X-Forwarded-Proto, X-Forwarded-Host, X-Forwarded-Prefix e.g. /api). Override anytime with OPENAPI_PUBLIC_BASE_URL.
func RegisterOpenAPISpec(r interface {
	Get(pattern string, h http.HandlerFunc)
}, cfg *config.Config, spec []byte) {
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		base := openAPIPublicBase(cfg, r)
		out := patchOpenAPIServers(spec, base)
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(out)
	})
	r.Get("/docs", swaggerUIHandler())
}

// openAPIPublicBase returns the absolute public base URL without a trailing slash, or "" to keep embedded relative "/".
func openAPIPublicBase(cfg *config.Config, r *http.Request) string {
	if cfg != nil {
		if v := strings.TrimSpace(cfg.OpenAPIPublicBase); v != "" {
			return strings.TrimSuffix(v, "/")
		}
	}
	host := r.Host
	if host == "" {
		return ""
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	// Trust forwarded headers when a reverse proxy sets them (TLS termination, path prefix).
	if r.Header.Get("X-Forwarded-Proto") != "" || r.Header.Get("X-Forwarded-Host") != "" {
		if p := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); p != "" {
			proto = p
		}
		if h := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); h != "" {
			host = h
		}
	}
	prefix := strings.TrimSpace(r.Header.Get("X-Forwarded-Prefix"))
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return proto + "://" + host + prefix
}

func patchOpenAPIServers(spec []byte, base string) []byte {
	if base == "" {
		return spec
	}
	if !bytes.Contains(spec, []byte(openAPIServersRelative)) {
		return spec
	}
	repl := fmt.Appendf(nil, "servers:\n  - url: %q\n    description: Public base URL (OPENAPI_PUBLIC_BASE_URL or request-derived)\n", base)
	return bytes.Replace(spec, []byte(openAPIServersRelative), repl, 1)
}

func swaggerUIHandler() http.HandlerFunc {
	const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Internal Comm API — Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
  <style>body { margin: 0; }</style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin="anonymous"></script>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js" crossorigin="anonymous"></script>
<script>
  (function () {
    var specUrl = new URL("openapi.yaml", window.location.href).href;
    window.ui = SwaggerUIBundle({
      url: specUrl,
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      layout: "StandaloneLayout",
      tryItOutEnabled: true,
    });
  })();
</script>
</body>
</html>`
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

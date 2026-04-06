package httpserver

import (
	"net/http"
)

// RegisterOpenAPISpec registers GET /openapi.yaml and GET /docs (Swagger UI).
func RegisterOpenAPISpec(r interface {
	Get(pattern string, h http.HandlerFunc)
}, spec []byte) {
	r.Get("/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(spec)
	})
	r.Get("/docs", swaggerUIHandler())
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
  window.ui = SwaggerUIBundle({
    url: window.location.origin + "/openapi.yaml",
    dom_id: "#swagger-ui",
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    layout: "StandaloneLayout",
    tryItOutEnabled: true,
  });
</script>
</body>
</html>`
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

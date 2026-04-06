package apiembed

import _ "embed"

// OpenAPIYAML is the canonical OpenAPI 3 document (source file: openapi.yaml in this package).
//
//go:embed openapi.yaml
var OpenAPIYAML []byte

package httpserver

import (
	"testing"

	"github.com/chatchitganggang/internal-comm-backend/internal/config"
)

func TestMinioHealthURL(t *testing.T) {
	t.Parallel()
	if got := minioHealthURL(&config.Config{}); got != "" {
		t.Fatalf("empty endpoint: got %q", got)
	}
	if got := minioHealthURL(&config.Config{MinIOEndpoint: "localhost:9000"}); got != "http://localhost:9000/minio/health/live" {
		t.Fatalf("http: got %q", got)
	}
	if got := minioHealthURL(&config.Config{MinIOEndpoint: "https://s3.example.com", MinIOUseSSL: true}); got != "https://s3.example.com/minio/health/live" {
		t.Fatalf("strip + ssl: got %q", got)
	}
}

func TestKeycloakProbeURL(t *testing.T) {
	t.Parallel()
	if got := keycloakProbeURL(&config.Config{}); got != "" {
		t.Fatalf("empty: got %q", got)
	}
	if got := keycloakProbeURL(&config.Config{KeycloakIssuer: "http://localhost/realms/x"}); got != "http://localhost/realms/x" {
		t.Fatalf("issuer only: got %q", got)
	}
	if got := keycloakProbeURL(&config.Config{
		KeycloakIssuer:   "http://localhost/realms/x",
		KeycloakReadyURL: "http://kc:9000/health/ready",
	}); got != "http://kc:9000/health/ready" {
		t.Fatalf("ready wins: got %q", got)
	}
}

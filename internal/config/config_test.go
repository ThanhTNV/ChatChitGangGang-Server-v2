package config

import (
	"log/slog"
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("KEYCLOAK_ISSUER", "")
	t.Setenv("KEYCLOAK_AUDIENCE", "")
	t.Setenv("KEYCLOAK_JWKS_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTP_ADDR default: got %q", cfg.HTTPAddr)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("LOG_LEVEL default: got %v", cfg.LogLevel)
	}
	if cfg.LogJSON {
		t.Fatal("LOG_FORMAT default should be text")
	}
}

func TestLoad_jsonAndDebug(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("KEYCLOAK_ISSUER", "")
	t.Setenv("KEYCLOAK_AUDIENCE", "")
	t.Setenv("KEYCLOAK_JWKS_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":9090" || cfg.LogLevel != slog.LevelDebug || !cfg.LogJSON {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestLoad_databaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("KEYCLOAK_ISSUER", "")
	t.Setenv("KEYCLOAK_AUDIENCE", "")
	t.Setenv("KEYCLOAK_JWKS_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgres://u:p@localhost:5432/db" {
		t.Fatalf("DatabaseURL: %q", cfg.DatabaseURL)
	}
}

func TestLoad_invalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "nope")
	t.Setenv("KEYCLOAK_ISSUER", "")
	t.Setenv("KEYCLOAK_AUDIENCE", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_invalidLogFormat(t *testing.T) {
	t.Setenv("LOG_FORMAT", "xml")
	t.Setenv("KEYCLOAK_ISSUER", "")
	t.Setenv("KEYCLOAK_AUDIENCE", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_keycloakIssuerRequiresAudience(t *testing.T) {
	t.Setenv("KEYCLOAK_ISSUER", "http://localhost:8090/realms/demo")
	t.Setenv("KEYCLOAK_AUDIENCE", "")
	t.Setenv("KEYCLOAK_JWKS_URL", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when audience missing")
	}
}

func TestLoad_keycloakDerivesJWKSURL(t *testing.T) {
	t.Setenv("KEYCLOAK_ISSUER", "http://localhost:8090/realms/demo")
	t.Setenv("KEYCLOAK_AUDIENCE", "api")
	t.Setenv("KEYCLOAK_JWKS_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := "http://localhost:8090/realms/demo/protocol/openid-connect/certs"
	if cfg.KeycloakJWKSURL != want {
		t.Fatalf("JWKS URL: got %q want %q", cfg.KeycloakJWKSURL, want)
	}
}

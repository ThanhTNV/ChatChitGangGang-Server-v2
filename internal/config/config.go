package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Config holds process configuration loaded from the environment.
type Config struct {
	HTTPAddr           string
	DatabaseURL        string
	LogLevel           slog.Level
	LogJSON            bool
	KeycloakIssuer     string
	KeycloakAudience   string
	KeycloakJWKSURL    string
}

// Load reads configuration from the environment and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:         strings.TrimSpace(os.Getenv("HTTP_ADDR")),
		DatabaseURL:      strings.TrimSpace(os.Getenv("DATABASE_URL")),
		LogLevel:         slog.LevelInfo,
		KeycloakIssuer:   strings.TrimSpace(os.Getenv("KEYCLOAK_ISSUER")),
		KeycloakAudience: strings.TrimSpace(os.Getenv("KEYCLOAK_AUDIENCE")),
		KeycloakJWKSURL:  strings.TrimSpace(os.Getenv("KEYCLOAK_JWKS_URL")),
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	rawLevel := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	if rawLevel == "" {
		rawLevel = "info"
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(strings.ToLower(rawLevel))); err != nil {
		return nil, fmt.Errorf("LOG_LEVEL: %w", err)
	}
	cfg.LogLevel = lvl

	logFmt := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
	switch logFmt {
	case "", "text":
		cfg.LogJSON = false
	case "json":
		cfg.LogJSON = true
	default:
		return nil, fmt.Errorf("LOG_FORMAT must be text or json, got %q", os.Getenv("LOG_FORMAT"))
	}

	if cfg.KeycloakIssuer != "" {
		if cfg.KeycloakAudience == "" {
			return nil, fmt.Errorf("KEYCLOAK_AUDIENCE is required when KEYCLOAK_ISSUER is set")
		}
		if cfg.KeycloakJWKSURL == "" {
			cfg.KeycloakJWKSURL = strings.TrimSuffix(cfg.KeycloakIssuer, "/") + "/protocol/openid-connect/certs"
		}
	}

	return cfg, nil
}

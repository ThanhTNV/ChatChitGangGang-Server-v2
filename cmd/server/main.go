package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/channel"
	"github.com/chatchitganggang/internal-comm-backend/internal/config"
	"github.com/chatchitganggang/internal-comm-backend/internal/httpserver"
	"github.com/chatchitganggang/internal-comm-backend/internal/user"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := newLogger(cfg)

	var dbPool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		poolCtx, poolCancel := context.WithTimeout(context.Background(), 15*time.Second)
		p, err := pgxpool.New(poolCtx, cfg.DatabaseURL)
		poolCancel()
		if err != nil {
			return err
		}
		dbPool = p
		defer dbPool.Close()
		log.Info("database configured", "pool_max_conns", dbPool.Config().MaxConns)
	}

	var authDeps *httpserver.Auth
	if cfg.KeycloakIssuer != "" {
		if dbPool == nil {
			return fmt.Errorf("DATABASE_URL is required when KEYCLOAK_ISSUER is set")
		}
		users := user.NewRepository(dbPool)
		chans := channel.NewRepository(dbPool)
		jwks := auth.NewJWKSStore(cfg.KeycloakJWKSURL, nil)
		v := auth.NewValidator(jwks, cfg.KeycloakIssuer, cfg.KeycloakAudience, 30*time.Second)
		authDeps = &httpserver.Auth{
			Bearer:    auth.BearerMiddleware(v, users),
			Validator: v,
			Users:     users,
			Channels:  chans,
		}
		log.Info("keycloak auth enabled", "issuer", cfg.KeycloakIssuer, "jwks", cfg.KeycloakJWKSURL)
	}

	srv := httpserver.New(cfg, log, dbPool, authDeps)

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		if err := <-errCh; err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}

func newLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var h slog.Handler
	if cfg.LogJSON {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

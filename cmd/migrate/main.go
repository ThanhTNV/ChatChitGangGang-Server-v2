package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/chatchitganggang/internal-comm-backend/internal/dbmigrate"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var err error
	switch *direction {
	case "up":
		err = dbmigrate.Up(ctx, dsn)
	case "down":
		err = dbmigrate.Down(ctx, dsn)
	default:
		err = fmt.Errorf("unknown -direction %q", *direction)
	}
	if err != nil {
		slog.Error("migrate", "error", err)
		os.Exit(1)
	}
	slog.Info("migrate done", "direction", *direction)
}

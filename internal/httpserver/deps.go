package httpserver

import "github.com/redis/go-redis/v9"

// Deps holds optional runtime dependencies for routes beyond the DB pool and auth.
type Deps struct {
	// Redis is optional; when nil, /ready reports redis as skipped.
	Redis *redis.Client
}

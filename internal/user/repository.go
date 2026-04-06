package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists users keyed by Keycloak subject.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a user repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertByKeycloakSub inserts or updates a user row for the given Keycloak sub.
func (r *Repository) UpsertByKeycloakSub(ctx context.Context, keycloakSub, displayName string) (uuid.UUID, error) {
	if keycloakSub == "" {
		return uuid.Nil, fmt.Errorf("empty keycloak sub")
	}
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (keycloak_sub, display_name, updated_at)
		VALUES ($1, NULLIF($2, ''), now())
		ON CONFLICT (keycloak_sub) DO UPDATE
		SET display_name = COALESCE(EXCLUDED.display_name, users.display_name),
		    updated_at = now()
		RETURNING id
	`, keycloakSub, displayName).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert user: %w", err)
	}
	return id, nil
}

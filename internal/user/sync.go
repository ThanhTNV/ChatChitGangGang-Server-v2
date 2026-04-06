package user

import (
	"context"

	"github.com/google/uuid"
)

// Sync persists users keyed by Keycloak subject (implemented by Repository).
type Sync interface {
	UpsertByKeycloakSub(ctx context.Context, keycloakSub, displayName string) (uuid.UUID, error)
}

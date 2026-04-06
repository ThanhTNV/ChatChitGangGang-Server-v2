package channel

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Channel is a row the current user is a member of.
type Channel struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Lister returns channels visible to a user (implemented by Repository).
type Lister interface {
	ListForUser(ctx context.Context, userID uuid.UUID) ([]Channel, error)
}

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

// CreateChannelRequest is the JSON body for POST /v1/channels.
type CreateChannelRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Store lists and creates channels for authenticated users (implemented by Repository).
type Store interface {
	ListForUser(ctx context.Context, userID uuid.UUID) ([]Channel, error)
	// CreateGroup inserts a group channel and adds userID as an admin member (transactional).
	CreateGroup(ctx context.Context, userID uuid.UUID, name string) (Channel, error)
}

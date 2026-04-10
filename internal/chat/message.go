package chat

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotMember means the user is not in channel_members for the channel.
var ErrNotMember = errors.New("not a channel member")

// ErrInvalidCursor means the cursor query parameter could not be parsed.
var ErrInvalidCursor = errors.New("invalid cursor")

// Message is a row in `messages` exposed over REST.
type Message struct {
	ID        uuid.UUID       `json:"id"`
	ChannelID uuid.UUID       `json:"channel_id"`
	SenderID  uuid.UUID       `json:"sender_id"`
	Body      json.RawMessage `json:"body"`
	CreatedAt time.Time       `json:"created_at"`
	EditedAt  *time.Time      `json:"edited_at,omitempty"`
}

// ListMessagesOpts controls pagination for ListForMember.
type ListMessagesOpts struct {
	// Cursor is an opaque string from a previous response's next_cursor; empty for the first page.
	Cursor string
	// Limit is the max messages to return (clamped by the repository).
	Limit int
}

// ListMessagesResult is one page of messages (newest first).
type ListMessagesResult struct {
	Messages   []Message
	NextCursor string
}

// Store loads messages for users who are channel members.
type Store interface {
	ListForMember(ctx context.Context, userID, channelID uuid.UUID, opts ListMessagesOpts) (ListMessagesResult, error)
}

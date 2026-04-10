package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMessagePageSize = 50
	maxMessagePageSize     = 100
)

// Repository loads messages from Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a chat/message repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListForMember returns messages in the channel if userID is a member, newest first.
func (r *Repository) ListForMember(ctx context.Context, userID, channelID uuid.UUID, opts ListMessagesOpts) (ListMessagesResult, error) {
	var member bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM channel_members
			WHERE channel_id = $1 AND user_id = $2
		)
	`, channelID, userID).Scan(&member)
	if err != nil {
		return ListMessagesResult{}, fmt.Errorf("channel membership: %w", err)
	}
	if !member {
		return ListMessagesResult{}, ErrNotMember
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = defaultMessagePageSize
	}
	if limit > maxMessagePageSize {
		limit = maxMessagePageSize
	}
	fetch := limit + 1

	var rows pgx.Rows
	if opts.Cursor == "" {
		rows, err = r.pool.Query(ctx, `
			SELECT m.id, m.channel_id, m.sender_id, m.body, m.created_at, m.edited_at
			FROM messages m
			WHERE m.channel_id = $1
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT $2
		`, channelID, fetch)
	} else {
		t, id, decErr := decodeMessageCursor(opts.Cursor)
		if decErr != nil {
			return ListMessagesResult{}, fmt.Errorf("%w: %v", ErrInvalidCursor, decErr)
		}
		rows, err = r.pool.Query(ctx, `
			SELECT m.id, m.channel_id, m.sender_id, m.body, m.created_at, m.edited_at
			FROM messages m
			WHERE m.channel_id = $1
			  AND (m.created_at, m.id) < ($2::timestamptz, $3::uuid)
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT $4
		`, channelID, t, id, fetch)
	}
	if err != nil {
		return ListMessagesResult{}, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var out []Message
	for rows.Next() {
		var m Message
		var body []byte
		var editedAt *time.Time
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.SenderID, &body, &m.CreatedAt, &editedAt); err != nil {
			return ListMessagesResult{}, fmt.Errorf("scan message: %w", err)
		}
		if len(body) > 0 {
			m.Body = json.RawMessage(append([]byte(nil), body...))
		} else {
			m.Body = json.RawMessage("{}")
		}
		m.EditedAt = editedAt
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return ListMessagesResult{}, fmt.Errorf("messages rows: %w", err)
	}

	result := ListMessagesResult{Messages: out}
	if len(out) > limit {
		result.Messages = out[:limit]
		last := result.Messages[len(result.Messages)-1]
		nc, encErr := encodeMessageCursor(last.CreatedAt, last.ID)
		if encErr != nil {
			return ListMessagesResult{}, fmt.Errorf("encode next cursor: %w", encErr)
		}
		result.NextCursor = nc
	}
	if result.Messages == nil {
		result.Messages = []Message{}
	}
	return result, nil
}

// Ensure Repository implements Store.
var _ Store = (*Repository)(nil)

package channel

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository loads channel data from Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a channel repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListForUser returns channels where the user is a member, newest first.
func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Channel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name, c.type, c.created_by, c.created_at
		FROM channels c
		INNER JOIN channel_members m ON m.channel_id = c.id
		WHERE m.user_id = $1
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	var out []Channel
	for rows.Next() {
		var c Channel
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("channels rows: %w", err)
	}
	if out == nil {
		out = []Channel{}
	}
	return out, nil
}

// Ensure Repository implements Lister.
var _ Lister = (*Repository)(nil)

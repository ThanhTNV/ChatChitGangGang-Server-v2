package channel

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// CreateGroup inserts a `group` channel with created_by = userID and adds the user as `admin` in channel_members.
func (r *Repository) CreateGroup(ctx context.Context, userID uuid.UUID, name string) (Channel, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Channel{}, fmt.Errorf("begin channel create: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var c Channel
	err = tx.QueryRow(ctx, `
		INSERT INTO channels (name, type, created_by)
		VALUES ($1, 'group', $2)
		RETURNING id, name, type, created_by, created_at
	`, name, userID).Scan(&c.ID, &c.Name, &c.Type, &c.CreatedBy, &c.CreatedAt)
	if err != nil {
		return Channel{}, fmt.Errorf("insert channel: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO channel_members (channel_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, c.ID, userID)
	if err != nil {
		return Channel{}, fmt.Errorf("insert channel member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Channel{}, fmt.Errorf("commit channel create: %w", err)
	}
	return c, nil
}

// Ensure Repository implements Store.
var _ Store = (*Repository)(nil)

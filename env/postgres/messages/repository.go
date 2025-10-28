package messages

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles message data access operations
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new message repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// ListSent retrieves only sent messages from the database (where processed_at IS NOT NULL)
// If limit is 0, all sent messages are returned
func (r *Repository) ListSent(ctx context.Context, limit int) ([]*Message, error) {
	query := `
		SELECT id, phone_number, content, created_at, message_id, processed_at
		FROM messages
		WHERE processed_at IS NOT NULL
		ORDER BY created_at ASC
	`

	args := []interface{}{}
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.PhoneNumber,
			&msg.Content,
			&msg.CreatedAt,
			&msg.MessageID,
			&msg.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// ListAndLockUnsent retrieves unsent messages and locks them for processing
// Uses SELECT FOR UPDATE SKIP LOCKED to prevent multiple instances from processing the same messages
// This method MUST be called within a transaction
func (r *Repository) ListAndLockUnsent(ctx context.Context, tx pgx.Tx, limit int) ([]*Message, error) {
	query := `
		SELECT id, phone_number, content, created_at, message_id, processed_at
		FROM messages
		WHERE processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := tx.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query and lock unsent messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.PhoneNumber,
			&msg.Content,
			&msg.CreatedAt,
			&msg.MessageID,
			&msg.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// Create inserts a new message into the database
// The ID will be populated after successful insertion
func (r *Repository) Create(ctx context.Context, msg *Message) error {
	query := `
		INSERT INTO messages (phone_number, content, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		msg.PhoneNumber,
		msg.Content,
		msg.CreatedAt,
	).Scan(&msg.ID)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// UpdateWithTx modifies an existing message in the database within a transaction
// Only updates message_id and processed_at fields
func (r *Repository) UpdateWithTx(ctx context.Context, tx pgx.Tx, id int64, messageID *string, processedAt *time.Time) error {
	query := `
		UPDATE messages
		SET message_id = $1, processed_at = $2
		WHERE id = $3
	`

	result, err := tx.Exec(ctx, query, messageID, processedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("message with id %d not found", id)
	}

	return nil
}

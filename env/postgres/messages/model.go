package messages

import (
	"time"
)

// Message represents a message data model for PostgreSQL persistence
// This is a pure data structure with no business logic
type Message struct {
	ID          int64     `db:"id"`
	PhoneNumber string    `db:"phone_number"`
	Content     string    `db:"content"`
	CreatedAt   time.Time `db:"created_at"`

	MessageID   *string    `db:"message_id"`
	ProcessedAt *time.Time `db:"processed_at"`
}

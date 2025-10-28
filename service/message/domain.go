package message

import (
	"fmt"
	"regexp"
	"time"
)

// Message content constraints
const (
	MaxContentLength = 500
)

// phoneRegex validates international phone number format
var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// Message represents a message domain entity with business logic
type Message struct {
	ID          int64
	PhoneNumber string
	Content     string
	CreatedAt   time.Time

	MessageID   *string
	ProcessedAt *time.Time
}

// Validate checks if the message fields are valid
func (m *Message) Validate() error {
	// Validate phone number
	if m.PhoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}

	if !phoneRegex.MatchString(m.PhoneNumber) {
		return fmt.Errorf("invalid phone number format (expected: +1234567890)")
	}

	// Validate content
	if m.Content == "" {
		return fmt.Errorf("message content is required")
	}

	if len(m.Content) > MaxContentLength {
		return fmt.Errorf("message content exceeds maximum length of %d characters", MaxContentLength)
	}

	return nil
}

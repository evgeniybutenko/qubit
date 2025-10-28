package webhook

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Client manages webhook HTTP requests (fake implementation for testing)
type Client struct {
	webhookURL     string
	webhookAuthKey string
}

// NewClient creates a new webhook client
func NewClient(webhookURL, webhookAuthKey string) *Client {
	return &Client{
		webhookURL:     webhookURL,
		webhookAuthKey: webhookAuthKey,
	}
}

// SendMessage sends a message via the webhook (simulated)
// Waits 0-5 seconds and fails 20% of requests
func (c *Client) SendMessage(ctx context.Context, phoneNumber, content string) (string, error) {
	// Random timeout between 0 and 5 seconds
	timeoutDuration := time.Duration(rand.Intn(5000)) * time.Millisecond

	// Simulate the timeout
	select {
	case <-time.After(timeoutDuration):
		// Continue after timeout
	case <-ctx.Done():
		return "", fmt.Errorf("webhook call cancelled: %w", ctx.Err())
	}

	// 20% chance of failure
	if rand.Intn(100) < 20 {
		return "", fmt.Errorf("webhook call failed: random failure occurred")
	}

	// Return success with UUID
	return uuid.New().String(), nil
}

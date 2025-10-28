package messages

import (
	"time"

	"qubit/service/message"
)

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID          int64      `json:"id"`
	PhoneNumber string     `json:"phoneNumber"`
	Content     string     `json:"content"`
	CreatedAt   time.Time  `json:"createdAt"`
	MessageID   *string    `json:"messageId"`
	ProcessedAt *time.Time `json:"processedAt"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// SchedulerStatusResponse represents the scheduler status
type SchedulerStatusResponse struct {
	Running         bool    `json:"running"`
	Interval        string  `json:"interval"`
	IntervalMinutes float64 `json:"intervalMinutes"`
}

// MessageListResponse represents a list of messages
type MessageListResponse struct {
	Success  bool              `json:"success"`
	Count    int               `json:"count"`
	Messages []MessageResponse `json:"messages"`
}

// ToMessageResponse converts a domain message.Message to MessageResponse
func ToMessageResponse(msg *message.Message) MessageResponse {
	resp := MessageResponse{
		ID:          msg.ID,
		PhoneNumber: msg.PhoneNumber,
		Content:     msg.Content,
		CreatedAt:   msg.CreatedAt,
		MessageID:   msg.MessageID,
		ProcessedAt: msg.ProcessedAt,
	}

	return resp
}

// ToMessageResponseList converts a slice of domain messages to MessageResponse slice
func ToMessageResponseList(messages []*message.Message) []MessageResponse {
	if messages == nil {
		return []MessageResponse{}
	}

	responses := make([]MessageResponse, 0, len(messages))
	for _, msg := range messages {
		responses = append(responses, ToMessageResponse(msg))
	}

	return responses
}

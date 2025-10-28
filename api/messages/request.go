package messages

// CreateMessageRequest represents the request to create a new message
type CreateMessageRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Content     string `json:"content" binding:"required,max=500"`
}

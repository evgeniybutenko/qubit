package messages

import (
	"net/http"

	"qubit/service/message"

	"github.com/gin-gonic/gin"
)

// Handler handles message-related HTTP requests
type Handler struct {
	messageService *message.Service
}

// NewHandler creates a new message handler
func NewHandler(messageService *message.Service) *Handler {
	return &Handler{
		messageService: messageService,
	}
}

// GetSentMessages handles GET /messages
// @Summary Get all sent messages
// @Description Returns a list of all sent messages
// @Tags Messages
// @Produce json
// @Success 200 {object} dto.MessageListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /messages [get]
func (h *Handler) GetSentMessages(c *gin.Context) {
	messages, err := h.messageService.GetSentMessages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve sent messages: " + err.Error(),
		})
		return
	}

	// Convert to response DTOs
	messageResponses := ToMessageResponseList(messages)

	c.JSON(http.StatusOK, MessageListResponse{
		Success:  true,
		Count:    len(messageResponses),
		Messages: messageResponses,
	})
}

// CreateMessage handles POST /messages
// @Summary Create a new message
// @Description Creates a new message to be sent
// @Tags Messages
// @Accept json
// @Produce json
// @Param message body dto.CreateMessageRequest true "Message data"
// @Success 201 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /messages [post]
func (h *Handler) CreateMessage(c *gin.Context) {
	var req CreateMessageRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Create message
	message, err := h.messageService.CreateMessage(c.Request.Context(), req.PhoneNumber, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to create message: " + err.Error(),
		})
		return
	}

	messageResponse := ToMessageResponse(message)

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "Message created successfully",
		Data:    messageResponse,
	})
}

// Start handles POST /scheduler/start
// @Summary Start the message scheduler
// @Description Starts the automatic message sending scheduler
// @Tags Scheduler
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /scheduler/start [post]
func (h *Handler) Start(c *gin.Context) {
	// Get configuration from context or use defaults
	// For now, we'll hardcode reasonable defaults that match the config
	intervalMinutes := 2
	batchSize := 2

	err := h.messageService.StartScheduler(intervalMinutes, batchSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Scheduler started successfully",
	})
}

// Stop handles POST /scheduler/stop
// @Summary Stop the message scheduler
// @Description Stops the automatic message sending scheduler
// @Tags Scheduler
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /scheduler/stop [post]
func (h *Handler) Stop(c *gin.Context) {
	err := h.messageService.StopScheduler()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Scheduler stopped successfully",
	})
}

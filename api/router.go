package api

import (
	"github.com/gin-gonic/gin"

	"qubit/api/messages"
	"qubit/service/message"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(messageService *message.Service) *gin.Engine {
	messagesHandler := messages.NewHandler(messageService)

	// Set Gin to release mode for production
	// gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Apply global middleware
	router.Use(Recovery())
	router.Use(Logger())
	router.Use(CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "qubit-message-service",
		})
	})

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Message endpoints
		messages := v1.Group("/messages")
		{
			messages.GET("/", messagesHandler.GetSentMessages)
			messages.POST("", messagesHandler.CreateMessage)
		}

		// Scheduler endpoints
		scheduler := v1.Group("/scheduler")
		{
			scheduler.POST("/start", messagesHandler.Start)
			scheduler.POST("/stop", messagesHandler.Stop)
		}
	}

	return router
}

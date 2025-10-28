package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qubit/api"
	"qubit/env/config"
	"qubit/env/postgres"
	"qubit/env/webhook"
	"qubit/service/message"
)

func main() {
	log.Println("Starting Qubit Message Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("✓ Configuration loaded")

	// Initialize context
	ctx := context.Background()

	// Initialize PostgreSQL client
	postgresClient, err := postgres.NewClient(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresClient.Close()

	// Initialize webhook client
	webhookClient := webhook.NewClient(cfg.WebhookURL, cfg.WebhookAuthKey)

	log.Println("✓ Environment initialized")

	// Initialize services
	messageService := message.NewService(postgresClient, webhookClient, cfg.SchedulerIntervalMinutes, cfg.MessageBatchSize)

	log.Println("✓ Services initialized")

	// Setup router (handlers are initialized inside)
	router := api.SetupRouter(messageService)
	log.Println("✓ Router configured")

	// Start HTTP server in a goroutine
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting HTTP server on %s", serverAddr)

	go func() {
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("✓ Qubit Message Service is running!")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop scheduler gracefully
	if err := messageService.StopScheduler(); err != nil {
		log.Printf("Warning: failed to stop scheduler: %v", err)
	}

	// Give some time for cleanup
	time.Sleep(2 * time.Second)

	log.Println("✓ Server shutdown complete")
}

package message

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"qubit/env/postgres"
	"qubit/env/webhook"
	"qubit/pkg/scheduler"
)

// Service handles the business logic for message operations
type Service struct {
	postgres      *postgres.Client
	webhookClient *webhook.Client
	scheduler     *scheduler.Client

	intervalMinutes  int
	messageBatchSize int

	mu sync.Mutex // Mutex to prevent concurrent processing within the same instance
}

// NewService creates a new message service and starts the scheduler
func NewService(
	postgresClient *postgres.Client,
	webhookClient *webhook.Client,
	intervalMinutes int,
	messageBatchSize int,
) *Service {
	s := &Service{
		postgres:         postgresClient,
		webhookClient:    webhookClient,
		scheduler:        scheduler.Run(),
		intervalMinutes:  intervalMinutes,
		messageBatchSize: messageBatchSize,
	}

	// Start the scheduler automatically
	task := func(ctx context.Context) error {
		return s.ProcessUnsentMessages(ctx, s.messageBatchSize)
	}

	if err := s.scheduler.Start(task, s.intervalMinutes); err != nil {
		log.Printf("Warning: failed to start scheduler: %v", err)
	} else {
		log.Printf("✓ Scheduler started (interval: %d minutes, batch size: %d)", intervalMinutes, messageBatchSize)
	}

	return s
}

// GetSentMessages retrieves all sent messages
func (s *Service) GetSentMessages(ctx context.Context) ([]*Message, error) {
	dbMessages, err := s.postgres.Messages.ListSent(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent messages: %w", err)
	}

	return ToDomainSlice(dbMessages), nil
}

// CreateMessage creates a new message
func (s *Service) CreateMessage(ctx context.Context, phoneNumber, content string) (*Message, error) {
	// Create domain message with validation
	msg := &Message{
		PhoneNumber: phoneNumber,
		Content:     content,
		CreatedAt:   time.Now(),
	}

	// Validate before inserting
	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Insert into database
	dbMsg := ToPostgres(msg)

	if err := s.postgres.Messages.Create(ctx, dbMsg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update domain model with generated ID
	msg.ID = dbMsg.ID

	return msg, nil
}

// ProcessUnsentMessages fetches and sends unsent messages
// This is the core function called by the scheduler
// Uses SELECT FOR UPDATE SKIP LOCKED to prevent duplicate processing across multiple instances
func (s *Service) ProcessUnsentMessages(ctx context.Context, batchSize int) error {
	// Lock to prevent concurrent processing within same instance
	s.mu.Lock()
	defer s.mu.Unlock()

	// Begin transaction
	tx, err := s.postgres.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is rolled back on error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Printf("Warning: failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	// Fetch and lock unsent messages atomically
	dbMessages, err := s.postgres.Messages.ListAndLockUnsent(ctx, tx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to fetch and lock unsent messages: %w", err)
	}

	if len(dbMessages) == 0 {
		// No messages to process, commit empty transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		log.Println("No unsent messages to process")
		return nil
	}

	log.Printf("Processing %d unsent messages (locked for this instance)", len(dbMessages))

	// Convert to domain models
	unsentMessages := ToDomainSlice(dbMessages)

	// Send each message and update within transaction
	for _, msg := range unsentMessages {
		if sendErr := s.sendMessageWithTx(ctx, tx, msg); sendErr != nil {
			log.Printf("Error sending message %d: %v", msg.ID, sendErr)
			// Continue processing other messages even if one fails
		}
	}

	// Commit transaction to release locks and persist updates
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✓ Batch processing complete, transaction committed")

	return nil
}

// sendMessageWithTx sends a single message and updates its status within a transaction
func (s *Service) sendMessageWithTx(ctx context.Context, tx pgx.Tx, msg *Message) error {
	log.Printf("Sending message %d to %s", msg.ID, msg.PhoneNumber)

	// Send message via webhook
	messageID, err := s.webhookClient.SendMessage(ctx, msg.PhoneNumber, msg.Content)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Mark as sent within the transaction
	sentAt := time.Now()
	err = s.postgres.Messages.UpdateWithTx(ctx, tx, msg.ID, &messageID, &sentAt)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	log.Printf("✓ Message %d sent successfully (messageId: %s)", msg.ID, messageID)

	return nil
}

// StartScheduler restarts the automatic message processing
func (s *Service) StartScheduler(intervalMinutes, batchSize int) error {
	if err := s.scheduler.Stop(); err != nil {
		log.Printf("Warning: failed to stop scheduler before restart: %v", err)
	}

	// Update configuration
	s.intervalMinutes = intervalMinutes
	s.messageBatchSize = batchSize

	// Start with new parameters
	task := func(ctx context.Context) error {
		return s.ProcessUnsentMessages(ctx, s.messageBatchSize)
	}

	return s.scheduler.Start(task, s.intervalMinutes)
}

// StopScheduler stops the automatic message processing
func (s *Service) StopScheduler() error {
	return s.scheduler.Stop()
}

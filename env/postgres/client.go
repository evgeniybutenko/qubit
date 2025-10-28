package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"qubit/env/postgres/messages"
)

// Client wraps the PostgreSQL connection pool and repositories
type Client struct {
	pool     *pgxpool.Pool
	Messages *messages.Repository
}

// NewClient creates a new PostgreSQL client with connection pool
func NewClient(ctx context.Context, databaseURL string) (*Client, error) {
	// Parse connection string and create pool config
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ PostgreSQL connection established successfully")

	client := &Client{
		pool:     pool,
		Messages: messages.NewRepository(pool),
	}

	return client, nil
}

// BeginTx starts a new database transaction
func (c *Client) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// Close gracefully closes the database connection pool
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
		log.Println("✓ PostgreSQL connection closed")
	}
}

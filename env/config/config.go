package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Database configuration
	DatabaseURL string

	// Webhook configuration
	WebhookURL     string
	WebhookAuthKey string

	// Server configuration
	ServerPort string

	// Scheduler configuration
	SchedulerIntervalMinutes int
	MessageBatchSize         int
}

// Load reads configuration from environment variables
// It automatically loads from .env file if present
func Load() (*Config, error) {
	// Load .env file if it exists. We ignore errors here because:
	// 1. In production, env vars are typically set in the system (Docker, K8s, etc.)
	// 2. The .env file is primarily for local development convenience
	// 3. Missing required variables will be caught by Validate() below
	// Note: This also silently ignores syntax errors in .env - if variables seem missing,
	// check your .env file for typos or malformed lines
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:              getEnv("DATABASE_URL", ""),
		WebhookURL:               getEnv("WEBHOOK_URL", ""),
		WebhookAuthKey:           getEnv("WEBHOOK_AUTH_KEY", ""),
		ServerPort:               getEnv("SERVER_PORT", "8080"),
		SchedulerIntervalMinutes: getEnvAsInt("SCHEDULER_INTERVAL_MINUTES", 2),
		MessageBatchSize:         getEnvAsInt("MESSAGE_BATCH_SIZE", 2),
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if all required configuration values are set
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.WebhookURL == "" {
		return fmt.Errorf("WEBHOOK_URL is required")
	}

	if c.WebhookAuthKey == "" {
		return fmt.Errorf("WEBHOOK_AUTH_KEY is required")
	}

	if c.SchedulerIntervalMinutes <= 0 {
		return fmt.Errorf("SCHEDULER_INTERVAL_MINUTES must be greater than 0")
	}

	if c.MessageBatchSize <= 0 {
		return fmt.Errorf("MESSAGE_BATCH_SIZE must be greater than 0")
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

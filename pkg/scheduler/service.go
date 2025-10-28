package scheduler

import (
	"context"
	"log"
	"sync"
	"time"
)

// Client manages the automatic task execution
type Client struct {
	task     func(context.Context) error
	interval time.Duration

	// Scheduler state
	ticker      *time.Ticker
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	wg          sync.WaitGroup
	taskRunning sync.Mutex // Prevents concurrent task executions
}

// Run starts a new scheduler client
func Run() *Client {
	return &Client{}
}

// Start starts the scheduler with the given task and interval
func (c *Client) Start(task func(context.Context) error, intervalMinutes int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.task = task
	c.interval = time.Duration(intervalMinutes) * time.Minute

	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.ticker = time.NewTicker(c.interval)

	c.wg.Add(1)
	go c.run()

	log.Printf("✓ Scheduler started (interval: %v)", c.interval)

	return nil
}

// Stop stops the scheduler gracefully
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("Stopping scheduler...")

	if c.ticker != nil {
		c.ticker.Stop()
	}

	if c.cancel != nil {
		c.cancel()
	}

	c.wg.Wait()

	log.Println("✓ Scheduler stopped")

	return nil
}

// run is the main scheduler loop
func (c *Client) run() {
	defer c.wg.Done()

	log.Println("Scheduler loop started")

	c.processTask()

	for {
		select {
		case <-c.ticker.C:
			c.processTask()

		case <-c.ctx.Done():
			log.Println("Scheduler context cancelled, exiting loop")
			return
		}
	}
}

// processTask executes the scheduled task
func (c *Client) processTask() {
	if !c.taskRunning.TryLock() {
		log.Println("⚠ Scheduler tick skipped: previous task still running")
		return
	}
	defer c.taskRunning.Unlock()

	log.Printf("--- Scheduler tick at %s ---", time.Now().Format(time.RFC3339))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := c.task(ctx)
	if err != nil {
		log.Printf("Error executing task: %v", err)
		return
	}

	log.Println("--- Scheduler tick complete ---")
}

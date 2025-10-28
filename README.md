# Qubit Message Service

Golang Software Engineer Test Assignment - An automatic message sending system built with clean architecture.

## Overview

A web application that automatically sends messages from a database queue via webhook at scheduled intervals. Built with clean architecture principles, Docker, PostgreSQL.

### Key Features

- Automatic Message Processing: Sends 2 unsent messages every 2 minutes (configurable)
- RESTful API: Create messages and manage scheduler
- Clean Architecture: Separation of concerns with clear layer boundaries
- Webhook Integration: Sends messages via external webhook service (mocked real service to provide better testing posibilities)
- Health Checks: Built-in health monitoring
- Docker Support: Fully containerized with Docker Compose

## Architecture

The project follows clean architecture principles with the following layers:

- api/ - HTTP API Layer (Handlers, Middleware, Router)
- service/ - Business Logic Layer
- env/ - Infrastructure Layer (DB, Config, Webhook, Migrations)
- pkg/ - Shared Libraries (Scheduler)

## Requirements

- **Go**: 1.23+
- **Docker**: Latest version
- **Docker Compose**: Latest version

## Quick Start

1. Clone the repository
2. Copy `.env.example` to `.env` and configure your webhook URL
3. Run `docker-compose up --build -d`
4. Access the API at `http://localhost:8080`

## API Endpoints

### Messages

- `POST /api/v1/messages` - Create a new message
- `GET /api/v1/messages` - Get all sent messages

### Scheduler

- `POST /api/v1/scheduler/start` - Start the scheduler
- `POST /api/v1/scheduler/stop` - Stop the scheduler

### Health

- `GET /health` - Health check endpoint

## Configuration

Copy `.env.example` to `.env` and configure:

### Application Configuration

- `DATABASE_URL` - PostgreSQL connection string
- `WEBHOOK_URL` - External webhook endpoint
- `WEBHOOK_AUTH_KEY` - Authentication key for webhook
- `SERVER_PORT` - HTTP server port (default: 8080)
- `SCHEDULER_INTERVAL_MINUTES` - Processing interval in minutes (default: 2)
- `MESSAGE_BATCH_SIZE` - Messages per batch (default: 2)

### PostgreSQL Configuration (Docker Compose)

- `POSTGRES_USER` - PostgreSQL username (default: qubit_user)
- `POSTGRES_PASSWORD` - PostgreSQL password
- `POSTGRES_DB` - Database name (default: qubit_db)
- `POSTGRES_PORT` - PostgreSQL port (default: 5432)

## Development

### Running Locally

```bash
go mod download
go run main.go
```

### Building

```bash
go build -o qubit .
```

### Testing

```bash
go test ./...
```

## How It Works

1. User creates messages via API
2. Messages stored in PostgreSQL with `processed_at = NULL`
3. Scheduler runs every 2 minutes
4. Fetches 2 unsent messages
5. Sends to webhook and updates status

## Concurrent Processing & Scalability

The application is designed to support **horizontal scaling** - you can run multiple instances simultaneously without message duplication or conflicts.

### FOR UPDATE SKIP LOCKED

The system uses PostgreSQL's `FOR UPDATE SKIP LOCKED` mechanism to prevent race conditions:

```sql
SELECT * FROM messages
WHERE processed_at IS NULL
ORDER BY created_at ASC
LIMIT 2
FOR UPDATE SKIP LOCKED;
```

**How it works:**

1. **Row-Level Locking**: When Instance A selects messages, those rows are locked within the transaction
2. **Skip Locked Rows**: Instance B automatically skips the locked rows and selects the next available messages
3. **No Waiting**: Instances never wait for each other - they immediately get different messages
4. **Transaction Safety**: Locks are released only after the transaction commits/rolls back

**Benefits:**

- **No Conflicts**: Multiple instances never process the same message
- **High Availability**: If one instance fails, others continue processing
- **Better Throughput**: More instances = more messages processed per minute
- **Zero Coordination**: No need for distributed locks or coordination services

**Example with 3 instances:**

```
Queue: [Msg1, Msg2, Msg3, Msg4, Msg5, Msg6]

Instance A: Locks & processes [Msg1, Msg2]
Instance B: Locks & processes [Msg3, Msg4]  (skips Msg1, Msg2 - they're locked)
Instance C: Locks & processes [Msg5, Msg6]  (skips Msg1-4 - they're locked)
```

All three instances work in parallel without any conflicts!

## Database Schema

```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    content VARCHAR(500) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    message_id TEXT,
    processed_at TIMESTAMP
);
```

## Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

## License

This is a test assignment project.

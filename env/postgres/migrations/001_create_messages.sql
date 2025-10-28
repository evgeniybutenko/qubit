-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    content VARCHAR(500) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    message_id TEXT,
    processed_at TIMESTAMP
);

-- Create index on processed_at for faster queries
CREATE INDEX IF NOT EXISTS idx_messages_processed_at ON messages(processed_at) WHERE processed_at IS NULL;

-- Create index on created_at for efficient ordering
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
-- Migration: Create gmail_messages table for caching Gmail message summaries and content
CREATE TABLE IF NOT EXISTS gmail_messages (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    gmail_message_id VARCHAR(255) NOT NULL,
    thread_id VARCHAR(255),
    subject TEXT,
    sender TEXT,
    recipient TEXT,
    snippet TEXT,
    body TEXT,
    internal_date BIGINT,
    history_id BIGINT,
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_fetched_at TIMESTAMPTZ,
    -- AI categorization fields
    category VARCHAR(64),
    categorization_confidence FLOAT,
    -- Gmail message JSON (for future-proofing, optional)
    raw_json JSONB,
    UNIQUE(user_id, gmail_message_id)
);

CREATE INDEX IF NOT EXISTS idx_gmail_messages_user_msg ON gmail_messages(user_id, gmail_message_id);

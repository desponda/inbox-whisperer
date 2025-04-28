-- Inbox Whisperer: Consolidated initial schema migration (2025-04-22)

-- 1. Users
DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deactivated BOOLEAN NOT NULL DEFAULT FALSE
);

-- 1a. User Identities (external auth providers)
DROP TABLE IF EXISTS user_identities CASCADE;
CREATE TABLE user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL, -- e.g. 'google', 'github'
    provider_user_id VARCHAR(128) NOT NULL, -- e.g. Google user id
    email VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider, provider_user_id),
    UNIQUE(user_id, provider)
);

-- 2. Categories
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- 3. Emails (Gmail, provider-agnostic)
DROP TABLE IF EXISTS emails CASCADE;
CREATE TABLE emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    gmail_id VARCHAR(128) NOT NULL,
    subject TEXT,
    sender TEXT,
    received_at TIMESTAMP WITH TIME ZONE,
    raw_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, gmail_id)
);

-- 4. EmailCategoryAssignments
CREATE TABLE IF NOT EXISTS email_category_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_id UUID REFERENCES emails(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    status VARCHAR(32),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(email_id, category_id)
);

-- 5. ActionLogs
DROP TABLE IF EXISTS action_logs CASCADE;
CREATE TABLE action_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(64),
    target_id INTEGER,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. EmailMessages (provider-agnostic, matches Go model)
DROP TABLE IF EXISTS email_messages CASCADE;
CREATE TABLE email_messages (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    email_message_id TEXT NOT NULL,
    thread_id TEXT,
    subject TEXT,
    sender TEXT,
    recipient TEXT,
    snippet TEXT,
    body TEXT,
    html_body TEXT,
    internal_date BIGINT,
    history_id BIGINT,
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_fetched_at TIMESTAMP,
    category TEXT,
    categorization_confidence FLOAT,
    raw_json JSONB,
    UNIQUE(user_id, provider, email_message_id)
);

CREATE INDEX IF NOT EXISTS idx_email_messages_user_id ON email_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_email_messages_email_message_id ON email_messages(email_message_id);

-- 7. UserTokens
DROP TABLE IF EXISTS user_tokens CASCADE;
CREATE TABLE user_tokens (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    token_json TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- 8. Seed categories
INSERT INTO categories (name, description) VALUES
    ('Primary', 'Personal and important emails'),
    ('Promotions/Ads', 'Marketing and promotional emails'),
    ('Social', 'Social network notifications'),
    ('Updates', 'Bills, receipts, confirmations'),
    ('Forums', 'Mailing lists and forums'),
    ('To Review', 'Updates, reports, notifications, no response needed'),
    ('Important', 'Personal, time-sensitive, or requires action'),
    ('Deferred', 'Requires action or follow-up later')
ON CONFLICT (name) DO NOTHING;

-- 9. Indexes
CREATE INDEX IF NOT EXISTS idx_emails_user_id ON emails(user_id);
CREATE INDEX IF NOT EXISTS idx_email_assignments_email_id ON email_category_assignments(email_id);
CREATE INDEX IF NOT EXISTS idx_action_logs_user_id ON action_logs(user_id);

-- 10. Sessions
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    values JSONB NOT NULL
);

-- Index for faster cleanup of expired sessions
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Index for looking up sessions by user
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);

-- Add comment to table
COMMENT ON TABLE sessions IS 'Stores user session data';

-- Ensure expired sessions are cleaned up
DELETE FROM sessions WHERE expires_at <= NOW();

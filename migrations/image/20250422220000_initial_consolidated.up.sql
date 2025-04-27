-- Inbox Whisperer: Consolidated initial schema migration (2025-04-22)

-- 1. Users
-- TEMPORARY: Using TEXT for user id (Google user id) for development only. See docs/features/future/identity-refactor.md for tech debt plan.
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deactivated BOOLEAN NOT NULL DEFAULT FALSE
);

-- 2. Categories
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- 3. Emails (Gmail, provider-agnostic)
CREATE TABLE IF NOT EXISTS emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
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
CREATE TABLE IF NOT EXISTS action_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(64),
    target_id INTEGER,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. EmailMessages (provider-agnostic, matches Go model)
CREATE TABLE IF NOT EXISTS email_messages (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
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
    UNIQUE(user_id, email_message_id)
);

CREATE INDEX IF NOT EXISTS idx_email_messages_user_id ON email_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_email_messages_email_message_id ON email_messages(email_message_id);

-- 7. UserTokens
CREATE TABLE IF NOT EXISTS user_tokens (
    user_id TEXT PRIMARY KEY,
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

-- Inbox Whisperer: Initial schema migration (see schema.sql for details)

-- 1. Users: App users (multi-user support)
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Categories: AI-driven email categories
CREATE TABLE IF NOT EXISTS categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- 3. Emails: Individual emails fetched from Gmail
CREATE TABLE IF NOT EXISTS emails (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER REFERENCES users(id) ON DELETE CASCADE,
    gmail_id        VARCHAR(128) NOT NULL, -- Gmail's unique email ID
    subject         TEXT,
    sender          TEXT,
    received_at     TIMESTAMP WITH TIME ZONE,
    raw_metadata    JSONB, -- Flexible storage for all Gmail metadata
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, gmail_id) -- Prevent duplicates per user
);

-- 4. EmailCategoryAssignments: Tracks category/status of each email
CREATE TABLE IF NOT EXISTS email_category_assignments (
    id              SERIAL PRIMARY KEY,
    email_id        INTEGER REFERENCES emails(id) ON DELETE CASCADE,
    category_id     INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    status          VARCHAR(32), -- e.g., 'inbox', 'archived', 'spam', etc.
    assigned_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(email_id, category_id)
);

-- 5. ActionLogs: Tracks user/system actions for audit/history
CREATE TABLE IF NOT EXISTS action_logs (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action          VARCHAR(64) NOT NULL,
    target_type     VARCHAR(64),
    target_id       INTEGER,
    details         JSONB,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed categories (idempotent)
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

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_emails_user_id ON emails(user_id);
CREATE INDEX IF NOT EXISTS idx_email_assignments_email_id ON email_category_assignments(email_id);
CREATE INDEX IF NOT EXISTS idx_action_logs_user_id ON action_logs(user_id);
CREATE INDEX idx_action_logs_user_id ON action_logs(user_id);

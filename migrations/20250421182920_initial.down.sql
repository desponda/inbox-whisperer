-- Rollback for initial Inbox Whisperer schema
DROP INDEX IF EXISTS idx_action_logs_user_id;
DROP INDEX IF EXISTS idx_email_assignments_email_id;
DROP INDEX IF EXISTS idx_emails_user_id;
DROP TABLE IF EXISTS action_logs;
DROP TABLE IF EXISTS email_category_assignments;
DROP TABLE IF EXISTS emails;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;

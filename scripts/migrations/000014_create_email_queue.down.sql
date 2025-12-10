-- Rollback: Drop email_queue table
-- Feature: M14-email-notifications

DROP INDEX IF EXISTS idx_email_queue_failed;
DROP INDEX IF EXISTS idx_email_queue_recipient;
DROP INDEX IF EXISTS idx_email_queue_pending;
DROP TABLE IF EXISTS email_queue;

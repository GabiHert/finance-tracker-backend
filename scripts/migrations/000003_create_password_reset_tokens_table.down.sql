-- Migration: Drop Password Reset Tokens Table
-- Description: Removes the password_reset_tokens table (DESTRUCTIVE)
-- Version: 000003

DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;
DROP INDEX IF EXISTS idx_password_reset_tokens_token;
DROP TABLE IF EXISTS password_reset_tokens;

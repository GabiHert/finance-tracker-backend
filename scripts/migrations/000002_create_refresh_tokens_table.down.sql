-- Migration: Drop Refresh Tokens Table
-- Description: Removes the refresh_tokens table (DESTRUCTIVE)
-- Version: 000002

DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP TABLE IF EXISTS refresh_tokens;

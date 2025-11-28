-- Migration: Create Refresh Tokens Table
-- Description: Creates the refresh_tokens table for JWT token management
-- Version: 000002
-- Date: 2024-01-01

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(500) NOT NULL,
    user_id UUID NOT NULL,
    invalidated BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create unique index on token for lookup
CREATE UNIQUE INDEX idx_refresh_tokens_token ON refresh_tokens(token);

-- Create index on user_id for user session queries
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Create index on expires_at for cleanup queries
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Add comments for documentation
COMMENT ON TABLE refresh_tokens IS 'Stores JWT refresh tokens for session management';
COMMENT ON COLUMN refresh_tokens.token IS 'Hashed refresh token value';
COMMENT ON COLUMN refresh_tokens.invalidated IS 'Flag to mark token as invalidated (logout)';
COMMENT ON COLUMN refresh_tokens.expires_at IS 'Token expiration timestamp (7 days default, 30 days with remember_me)';

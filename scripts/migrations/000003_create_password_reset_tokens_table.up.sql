-- Migration: Create Password Reset Tokens Table
-- Description: Creates the password_reset_tokens table for password recovery
-- Version: 000003
-- Date: 2024-01-01

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(500) NOT NULL,
    user_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_password_reset_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create unique index on token for lookup
CREATE UNIQUE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);

-- Create index on user_id for user queries
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);

-- Create index on expires_at for cleanup queries
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- Add comments for documentation
COMMENT ON TABLE password_reset_tokens IS 'Stores single-use password reset tokens';
COMMENT ON COLUMN password_reset_tokens.token IS 'Hashed reset token value';
COMMENT ON COLUMN password_reset_tokens.used IS 'Flag indicating token has been consumed';
COMMENT ON COLUMN password_reset_tokens.used_at IS 'Timestamp when token was used';
COMMENT ON COLUMN password_reset_tokens.expires_at IS 'Token expiration (1 hour from creation)';

-- Migration: Create Group Invites Table
-- Description: Creates the group_invites table for managing group invitations
-- Version: 000008
-- Date: 2025-11-29
-- Milestone: M9 - Groups & Collaboration

-- Create invite_status enum
CREATE TYPE invite_status AS ENUM ('pending', 'accepted', 'declined', 'expired');

CREATE TABLE IF NOT EXISTS group_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status invite_status NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_group_invites_group_id ON group_invites(group_id);
CREATE INDEX idx_group_invites_email ON group_invites(email);
CREATE INDEX idx_group_invites_token ON group_invites(token);
CREATE INDEX idx_group_invites_status ON group_invites(status);
CREATE INDEX idx_group_invites_expires_at ON group_invites(expires_at);

-- Unique constraint: only one pending invite per email per group
CREATE UNIQUE INDEX idx_group_invites_pending ON group_invites(group_id, email) WHERE status = 'pending';

-- Add comments for documentation
COMMENT ON TABLE group_invites IS 'Stores pending and historical group invitations';
COMMENT ON COLUMN group_invites.email IS 'Email address of the invited person';
COMMENT ON COLUMN group_invites.token IS 'Unique token for accepting the invitation';
COMMENT ON COLUMN group_invites.invited_by IS 'UUID of the user who sent the invitation';
COMMENT ON COLUMN group_invites.status IS 'Invitation status: pending, accepted, declined, or expired';
COMMENT ON COLUMN group_invites.expires_at IS 'When the invitation expires';

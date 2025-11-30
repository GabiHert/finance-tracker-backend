-- Migration: Create Group Members Table
-- Description: Creates the group_members table for tracking group membership
-- Version: 000007
-- Date: 2025-11-29
-- Milestone: M9 - Groups & Collaboration

-- Create member_role enum
CREATE TYPE member_role AS ENUM ('admin', 'member');

CREATE TABLE IF NOT EXISTS group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role member_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_group_member UNIQUE (group_id, user_id)
);

-- Create indexes for efficient querying
CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE INDEX idx_group_members_role ON group_members(role);

-- Add comments for documentation
COMMENT ON TABLE group_members IS 'Stores membership relationships between users and groups';
COMMENT ON COLUMN group_members.group_id IS 'Reference to the group';
COMMENT ON COLUMN group_members.user_id IS 'Reference to the member user';
COMMENT ON COLUMN group_members.role IS 'Member role: admin or member';
COMMENT ON COLUMN group_members.joined_at IS 'When the user joined the group';

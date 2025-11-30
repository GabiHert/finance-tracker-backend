-- Migration: Create Groups Table
-- Description: Creates the groups table for collaborative finance management
-- Version: 000006
-- Date: 2025-11-29
-- Milestone: M9 - Groups & Collaboration

CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_groups_created_by ON groups(created_by);
CREATE INDEX idx_groups_name ON groups(name);

-- Add comments for documentation
COMMENT ON TABLE groups IS 'Stores collaborative groups for shared finance management';
COMMENT ON COLUMN groups.name IS 'Display name of the group (e.g., Familia Silva)';
COMMENT ON COLUMN groups.created_by IS 'UUID of the user who created the group';

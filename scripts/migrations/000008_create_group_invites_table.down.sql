-- Migration: Drop Group Invites Table
-- Description: Drops the group_invites table and invite_status enum
-- Version: 000008
-- Date: 2025-11-29

DROP TABLE IF EXISTS group_invites CASCADE;
DROP TYPE IF EXISTS invite_status;

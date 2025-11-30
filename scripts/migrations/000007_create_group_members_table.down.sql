-- Migration: Drop Group Members Table
-- Description: Drops the group_members table and member_role enum
-- Version: 000007
-- Date: 2025-11-29

DROP TABLE IF EXISTS group_members CASCADE;
DROP TYPE IF EXISTS member_role;

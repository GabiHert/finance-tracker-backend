-- Migration: Drop Users Table
-- Description: Removes the users table (DESTRUCTIVE)
-- Version: 000001

DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;

-- Migration: Create Users Table
-- Description: Creates the users table for authentication and user management
-- Version: 000001
-- Date: 2024-01-01

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    date_format VARCHAR(20) NOT NULL DEFAULT 'YYYY-MM-DD',
    number_format VARCHAR(10) NOT NULL DEFAULT 'US',
    first_day_of_week VARCHAR(10) NOT NULL DEFAULT 'sunday',
    email_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    goal_alerts BOOLEAN NOT NULL DEFAULT TRUE,
    recurring_reminders BOOLEAN NOT NULL DEFAULT TRUE,
    terms_accepted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create unique index on email for login lookups
CREATE UNIQUE INDEX idx_users_email ON users(email);

-- Add comment for table documentation
COMMENT ON TABLE users IS 'Stores user account information for authentication and preferences';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password with cost factor 12';
COMMENT ON COLUMN users.date_format IS 'User preferred date format (ISO 8601)';
COMMENT ON COLUMN users.number_format IS 'User preferred number format (US, EU, etc)';
COMMENT ON COLUMN users.first_day_of_week IS 'User preferred first day of week (sunday, monday)';

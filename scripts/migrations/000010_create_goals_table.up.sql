-- Migration: Create Goals Table
-- Description: Creates the goals table for spending limits management
-- Version: 000010
-- Date: 2025-12-01
-- Milestone: M7 - Goals (Spending Limits) API

-- Create period_type enum for goal periods
CREATE TYPE goal_period_type AS ENUM ('monthly', 'weekly', 'yearly');

CREATE TABLE IF NOT EXISTS goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    limit_amount DECIMAL(15,2) NOT NULL,
    alert_on_exceed BOOLEAN NOT NULL DEFAULT true,
    period goal_period_type NOT NULL DEFAULT 'monthly',
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create unique index for user_id + category_id (only for non-deleted records)
CREATE UNIQUE INDEX idx_goals_user_category ON goals(user_id, category_id) WHERE deleted_at IS NULL;

-- Create indexes for efficient querying
CREATE INDEX idx_goals_user_id ON goals(user_id);
CREATE INDEX idx_goals_category_id ON goals(category_id);

-- Add comments for documentation
COMMENT ON TABLE goals IS 'Stores spending limit goals per category for users';
COMMENT ON COLUMN goals.user_id IS 'UUID of the user who owns this goal';
COMMENT ON COLUMN goals.category_id IS 'UUID of the category this goal applies to';
COMMENT ON COLUMN goals.limit_amount IS 'Maximum spending limit amount for the period';
COMMENT ON COLUMN goals.alert_on_exceed IS 'Whether to alert user when spending exceeds limit';
COMMENT ON COLUMN goals.period IS 'Time period for the goal: monthly, weekly, or yearly';
COMMENT ON COLUMN goals.start_date IS 'Optional start date for custom period goals';
COMMENT ON COLUMN goals.end_date IS 'Optional end date for custom period goals';
COMMENT ON COLUMN goals.deleted_at IS 'Soft delete timestamp';

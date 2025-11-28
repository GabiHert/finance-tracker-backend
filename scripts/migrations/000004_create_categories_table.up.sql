-- Migration: Create Categories Table
-- Description: Creates the categories table for transaction categorization
-- Version: 000004
-- Date: 2025-11-27
-- Milestone: M3 - Core Data Models & Categories

-- Create owner_type enum
CREATE TYPE owner_type AS ENUM ('user', 'group');

-- Create category_type enum
CREATE TYPE category_type AS ENUM ('expense', 'income');

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6366F1',
    icon VARCHAR(50) NOT NULL DEFAULT 'tag',
    owner_type owner_type NOT NULL,
    owner_id UUID NOT NULL,
    type category_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_categories_owner ON categories(owner_type, owner_id);
CREATE INDEX idx_categories_type ON categories(type);

-- Unique constraint: category name must be unique per owner
CREATE UNIQUE INDEX idx_categories_name_owner ON categories(name, owner_type, owner_id);

-- Add comments for documentation
COMMENT ON TABLE categories IS 'Stores transaction categories owned by users or groups';
COMMENT ON COLUMN categories.name IS 'Category display name (e.g., Food, Transport)';
COMMENT ON COLUMN categories.color IS 'Hex color code for UI display (e.g., #6366F1)';
COMMENT ON COLUMN categories.icon IS 'Icon identifier from icon library';
COMMENT ON COLUMN categories.owner_type IS 'Whether category belongs to user or group';
COMMENT ON COLUMN categories.owner_id IS 'UUID of owning user or group';
COMMENT ON COLUMN categories.type IS 'expense or income category type';

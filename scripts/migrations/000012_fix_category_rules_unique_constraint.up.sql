-- 000012_fix_category_rules_unique_constraint.up.sql
-- Fixes the unique constraint on category_rules to support soft-delete properly.
-- The previous constraint didn't exclude soft-deleted records, causing conflicts.

-- Add deleted_at column for soft-delete support (if not already added by GORM)
ALTER TABLE category_rules ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Create index on deleted_at for efficient soft-delete queries
CREATE INDEX IF NOT EXISTS idx_category_rules_deleted_at ON category_rules(deleted_at);

-- Drop the old unique constraint that includes all records
ALTER TABLE category_rules DROP CONSTRAINT IF EXISTS uq_category_rules_owner_pattern;

-- Create a partial unique index that excludes soft-deleted records
-- This allows the same pattern to be recreated after a rule is soft-deleted
CREATE UNIQUE INDEX idx_category_rules_owner_pattern_active
ON category_rules(owner_type, owner_id, pattern)
WHERE deleted_at IS NULL;

-- Comment for documentation
COMMENT ON INDEX idx_category_rules_owner_pattern_active IS 'Ensures unique patterns per owner for non-deleted rules only';

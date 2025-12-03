-- 000012_fix_category_rules_unique_constraint.down.sql
-- Reverts the unique constraint change on category_rules

-- Drop the partial unique index
DROP INDEX IF EXISTS idx_category_rules_owner_pattern_active;

-- Recreate the original unique constraint
ALTER TABLE category_rules ADD CONSTRAINT uq_category_rules_owner_pattern
UNIQUE (owner_type, owner_id, pattern);

-- Note: We don't drop the deleted_at column as it may have been added by GORM
-- and contains data we don't want to lose

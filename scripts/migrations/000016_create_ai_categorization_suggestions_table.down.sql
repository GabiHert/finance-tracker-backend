-- 000015_create_ai_categorization_suggestions_table.down.sql
-- Drops the ai_categorization_suggestions table

-- Drop indexes first
DROP INDEX IF EXISTS idx_ai_suggestions_user_status_created;
DROP INDEX IF EXISTS idx_ai_suggestions_deleted_at;
DROP INDEX IF EXISTS idx_ai_suggestions_category_id;
DROP INDEX IF EXISTS idx_ai_suggestions_transaction_id;
DROP INDEX IF EXISTS idx_ai_suggestions_status;
DROP INDEX IF EXISTS idx_ai_suggestions_user_id;

-- Drop table
DROP TABLE IF EXISTS ai_categorization_suggestions;

-- Drop enum types (optional, as they might be used elsewhere)
-- DROP TYPE IF EXISTS match_type;
-- DROP TYPE IF EXISTS suggestion_status;

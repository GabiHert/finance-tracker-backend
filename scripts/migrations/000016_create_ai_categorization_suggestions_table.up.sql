-- 000016_create_ai_categorization_suggestions_table.up.sql
-- Creates the ai_categorization_suggestions table for the AI Smart Categorization feature (M15)

-- Create enum types for suggestion status and match type
DO $$ BEGIN
    CREATE TYPE suggestion_status AS ENUM ('pending', 'approved', 'rejected', 'skipped');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE match_type AS ENUM ('contains', 'startsWith', 'exact');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create the ai_categorization_suggestions table
CREATE TABLE ai_categorization_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,                                -- Owner of the suggestion
    transaction_id UUID NOT NULL,                         -- Primary transaction that triggered suggestion
    suggested_category_id UUID,                           -- References categories(id), nullable for new category
    suggested_category_new JSONB,                         -- For new category: {name, icon, color}
    match_type VARCHAR(20) NOT NULL,                      -- 'contains', 'startsWith', 'exact'
    match_keyword VARCHAR(255) NOT NULL,                  -- The keyword/pattern to match
    affected_transaction_ids UUID[],                      -- Array of transaction IDs that would be affected
    status VARCHAR(20) NOT NULL DEFAULT 'pending',        -- 'pending', 'approved', 'rejected', 'skipped'
    previous_suggestion JSONB,                            -- JSON for retry context
    retry_reason TEXT,                                    -- User's reason for requesting retry
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,                  -- Soft-delete support

    -- Foreign key constraints
    CONSTRAINT fk_ai_suggestions_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_ai_suggestions_transaction FOREIGN KEY (transaction_id)
        REFERENCES transactions(id) ON DELETE CASCADE,
    CONSTRAINT fk_ai_suggestions_category FOREIGN KEY (suggested_category_id)
        REFERENCES categories(id) ON DELETE SET NULL,

    -- Check constraints
    CONSTRAINT chk_category_suggestion CHECK (
        (suggested_category_id IS NOT NULL AND suggested_category_new IS NULL) OR
        (suggested_category_id IS NULL AND suggested_category_new IS NOT NULL)
    ),
    CONSTRAINT chk_status CHECK (
        status IN ('pending', 'approved', 'rejected', 'skipped')
    ),
    CONSTRAINT chk_match_type CHECK (
        match_type IN ('contains', 'startsWith', 'exact')
    )
);

-- Index for user lookups (get suggestions for a user)
CREATE INDEX idx_ai_suggestions_user_id ON ai_categorization_suggestions(user_id);

-- Index for status lookups (get pending suggestions)
CREATE INDEX idx_ai_suggestions_status ON ai_categorization_suggestions(user_id, status);

-- Index for transaction lookups
CREATE INDEX idx_ai_suggestions_transaction_id ON ai_categorization_suggestions(transaction_id);

-- Index for category lookups
CREATE INDEX idx_ai_suggestions_category_id ON ai_categorization_suggestions(suggested_category_id);

-- Index for soft-delete queries
CREATE INDEX idx_ai_suggestions_deleted_at ON ai_categorization_suggestions(deleted_at) WHERE deleted_at IS NULL;

-- Composite index for common query pattern
CREATE INDEX idx_ai_suggestions_user_status_created ON ai_categorization_suggestions(user_id, status, created_at DESC)
    WHERE deleted_at IS NULL;

-- Comments for documentation
COMMENT ON TABLE ai_categorization_suggestions IS 'Stores AI-generated categorization suggestions for user review';
COMMENT ON COLUMN ai_categorization_suggestions.user_id IS 'Owner of the suggestion';
COMMENT ON COLUMN ai_categorization_suggestions.transaction_id IS 'Primary transaction that triggered the suggestion';
COMMENT ON COLUMN ai_categorization_suggestions.suggested_category_id IS 'ID of existing category if suggesting existing, NULL for new category';
COMMENT ON COLUMN ai_categorization_suggestions.suggested_category_new IS 'JSON object with new category details: {name, icon, color}';
COMMENT ON COLUMN ai_categorization_suggestions.match_type IS 'Type of pattern matching: contains, startsWith, or exact';
COMMENT ON COLUMN ai_categorization_suggestions.match_keyword IS 'Keyword/pattern to match for auto-categorization rule';
COMMENT ON COLUMN ai_categorization_suggestions.affected_transaction_ids IS 'Array of transaction IDs that would be affected by this rule';
COMMENT ON COLUMN ai_categorization_suggestions.status IS 'Status of the suggestion: pending, approved, rejected, or skipped';
COMMENT ON COLUMN ai_categorization_suggestions.previous_suggestion IS 'JSON context from previous suggestion if this is a retry';
COMMENT ON COLUMN ai_categorization_suggestions.retry_reason IS 'User-provided reason for requesting a new suggestion';

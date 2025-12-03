-- 000011_create_category_rules_table.up.sql
-- Creates the category_rules table for the Category Rules Engine (M6)

CREATE TABLE category_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern VARCHAR(255) NOT NULL,           -- Regex pattern
    category_id UUID NOT NULL,               -- References categories(id)
    priority INTEGER NOT NULL DEFAULT 0,     -- Higher = checked first
    is_active BOOLEAN NOT NULL DEFAULT true, -- Enable/disable without deleting
    owner_type VARCHAR(10) NOT NULL,         -- 'user' or 'group'
    owner_id UUID NOT NULL,                  -- user_id or group_id
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Foreign key constraint
    CONSTRAINT fk_category_rules_category FOREIGN KEY (category_id)
        REFERENCES categories(id) ON DELETE CASCADE,

    -- Unique constraint: no duplicate patterns per owner
    CONSTRAINT uq_category_rules_owner_pattern UNIQUE (owner_type, owner_id, pattern)
);

-- Index for efficient owner lookups
CREATE INDEX idx_category_rules_owner ON category_rules(owner_type, owner_id);

-- Index for priority-based ordering within an owner
CREATE INDEX idx_category_rules_priority ON category_rules(owner_type, owner_id, priority DESC);

-- Index for category lookups
CREATE INDEX idx_category_rules_category ON category_rules(category_id);

-- Comment on table and columns for documentation
COMMENT ON TABLE category_rules IS 'Stores auto-categorization rules based on regex patterns';
COMMENT ON COLUMN category_rules.pattern IS 'Regex pattern to match against transaction descriptions';
COMMENT ON COLUMN category_rules.priority IS 'Higher priority rules are checked first during auto-categorization';
COMMENT ON COLUMN category_rules.is_active IS 'Allows disabling rules without deleting them';
COMMENT ON COLUMN category_rules.owner_type IS 'Type of owner: user or group';
COMMENT ON COLUMN category_rules.owner_id IS 'ID of the owning user or group';

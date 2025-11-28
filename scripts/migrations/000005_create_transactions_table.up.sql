-- Migration: Create Transactions Table
-- Description: Creates the transactions table for financial transaction tracking
-- Version: 000005
-- Date: 2025-11-27
-- Milestone: M3 - Core Data Models & Categories

-- Create transaction_type enum (reusing category_type if compatible, else create new)
-- Note: Using category_type enum for expense/income since it's the same

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    description VARCHAR(500) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    type category_type NOT NULL,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    notes TEXT,
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
CREATE INDEX idx_transactions_type ON transactions(type);

-- Composite indexes for common query patterns
CREATE INDEX idx_transactions_user_date ON transactions(user_id, date);
CREATE INDEX idx_transactions_user_description ON transactions(user_id, description);

-- Add comments for documentation
COMMENT ON TABLE transactions IS 'Stores financial transactions imported or manually entered';
COMMENT ON COLUMN transactions.user_id IS 'Owner of the transaction';
COMMENT ON COLUMN transactions.date IS 'Transaction date';
COMMENT ON COLUMN transactions.description IS 'Merchant name or transaction description';
COMMENT ON COLUMN transactions.amount IS 'Amount in BRL (negative for expenses, positive for income)';
COMMENT ON COLUMN transactions.type IS 'expense or income (derived from amount sign)';
COMMENT ON COLUMN transactions.category_id IS 'Reference to category (nullable for uncategorized)';
COMMENT ON COLUMN transactions.notes IS 'Optional user notes';
COMMENT ON COLUMN transactions.is_recurring IS 'Recurring expense flag (system-detected, user-editable)';
COMMENT ON COLUMN transactions.uploaded_at IS 'When transaction was imported';

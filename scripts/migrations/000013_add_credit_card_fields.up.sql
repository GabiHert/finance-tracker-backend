-- Migration: Add credit card import fields to transactions
-- Feature: M12-cc-import
-- Description: Adds fields to support credit card statement import and bill matching

-- Add columns for credit card tracking
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS credit_card_payment_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS billing_cycle VARCHAR(7),
    ADD COLUMN IF NOT EXISTS original_amount DECIMAL(12, 2),
    ADD COLUMN IF NOT EXISTS is_credit_card_payment BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS expanded_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS installment_current INTEGER,
    ADD COLUMN IF NOT EXISTS installment_total INTEGER,
    ADD COLUMN IF NOT EXISTS is_hidden BOOLEAN DEFAULT false;

-- Indexes for credit card queries
CREATE INDEX IF NOT EXISTS idx_transactions_credit_card_payment_id
    ON transactions(credit_card_payment_id)
    WHERE credit_card_payment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_billing_cycle
    ON transactions(billing_cycle)
    WHERE billing_cycle IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_is_credit_card_payment
    ON transactions(is_credit_card_payment)
    WHERE is_credit_card_payment = true;

CREATE INDEX IF NOT EXISTS idx_transactions_is_hidden
    ON transactions(is_hidden)
    WHERE is_hidden = true;

-- Composite index for finding expandable bill payments
CREATE INDEX IF NOT EXISTS idx_transactions_user_cc_payment
    ON transactions(user_id, is_credit_card_payment, expanded_at)
    WHERE is_credit_card_payment = true;

-- Composite index for credit card status queries
CREATE INDEX IF NOT EXISTS idx_transactions_user_billing_cycle
    ON transactions(user_id, billing_cycle, credit_card_payment_id)
    WHERE billing_cycle IS NOT NULL;

-- Constraint: installment_current must be <= installment_total and > 0
ALTER TABLE transactions
    ADD CONSTRAINT chk_installment_valid
    CHECK (
        installment_current IS NULL
        OR (installment_current > 0 AND installment_total IS NOT NULL AND installment_current <= installment_total)
    );

-- Constraint: original_amount only set when is_credit_card_payment is true
ALTER TABLE transactions
    ADD CONSTRAINT chk_original_amount_valid
    CHECK (
        original_amount IS NULL
        OR is_credit_card_payment = true
    );

-- Constraint: expanded_at only set when is_credit_card_payment is true
ALTER TABLE transactions
    ADD CONSTRAINT chk_expanded_at_valid
    CHECK (
        expanded_at IS NULL
        OR is_credit_card_payment = true
    );

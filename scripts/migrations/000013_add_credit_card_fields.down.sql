-- Rollback: Remove credit card import fields from transactions
-- Feature: M12-cc-import

-- Drop constraints first
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_expanded_at_valid;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_original_amount_valid;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_installment_valid;

-- Drop indexes
DROP INDEX IF EXISTS idx_transactions_user_billing_cycle;
DROP INDEX IF EXISTS idx_transactions_user_cc_payment;
DROP INDEX IF EXISTS idx_transactions_is_hidden;
DROP INDEX IF EXISTS idx_transactions_is_credit_card_payment;
DROP INDEX IF EXISTS idx_transactions_billing_cycle;
DROP INDEX IF EXISTS idx_transactions_credit_card_payment_id;

-- Drop columns
ALTER TABLE transactions
    DROP COLUMN IF EXISTS is_hidden,
    DROP COLUMN IF EXISTS installment_total,
    DROP COLUMN IF EXISTS installment_current,
    DROP COLUMN IF EXISTS expanded_at,
    DROP COLUMN IF EXISTS is_credit_card_payment,
    DROP COLUMN IF EXISTS original_amount,
    DROP COLUMN IF EXISTS billing_cycle,
    DROP COLUMN IF EXISTS credit_card_payment_id;

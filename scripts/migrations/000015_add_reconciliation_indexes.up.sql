-- Migration: Add reconciliation indexes for M15-smart-reconciliation
-- Purpose: Optimize queries for pending CC transactions and bill matching

-- Index for finding pending (unlinked) CC transactions grouped by billing cycle
-- Used by: GetPendingBillingCycles query
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_pending_cc
ON transactions (user_id, billing_cycle)
WHERE billing_cycle IS NOT NULL
  AND credit_card_payment_id IS NULL
  AND deleted_at IS NULL
  AND is_hidden = false;

-- Index for finding potential bill payments in date range
-- Used by: FindPotentialBills query
-- Note: Using a simpler pattern check that PostgreSQL can optimize
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_potential_bills
ON transactions (user_id, date, amount)
WHERE type = 'expense'
  AND expanded_at IS NULL
  AND deleted_at IS NULL;

-- Index for checking if a bill is already linked
-- Used by: IsBillLinked query
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_cc_payment_ref
ON transactions (credit_card_payment_id)
WHERE credit_card_payment_id IS NOT NULL
  AND deleted_at IS NULL;

-- Composite index for efficient billing cycle aggregation
-- Used by: Dashboard pending indicator query
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_pending_summary
ON transactions (user_id, billing_cycle, amount)
WHERE billing_cycle IS NOT NULL
  AND credit_card_payment_id IS NULL
  AND deleted_at IS NULL
  AND is_hidden = false;

COMMENT ON INDEX idx_transactions_pending_cc IS 'M15: Find pending CC transactions by billing cycle';
COMMENT ON INDEX idx_transactions_potential_bills IS 'M15: Find potential bill payments for matching';
COMMENT ON INDEX idx_transactions_cc_payment_ref IS 'M15: Check if bill is already linked';
COMMENT ON INDEX idx_transactions_pending_summary IS 'M15: Dashboard pending reconciliation summary';

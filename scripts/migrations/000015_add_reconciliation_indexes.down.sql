-- Rollback: Remove reconciliation indexes

DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_pending_summary;
DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_cc_payment_ref;
DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_potential_bills;
DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_pending_cc;

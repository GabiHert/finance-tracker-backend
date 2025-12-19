// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// TransactionFilter defines filter options for listing transactions.
type TransactionFilter struct {
	UserID      uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	CategoryIDs []uuid.UUID
	Type        *entity.TransactionType
	Search      string // Case-insensitive description match
	GroupByDate bool
}

// TransactionPagination defines pagination options.
type TransactionPagination struct {
	Page  int
	Limit int
}

// TransactionListResult represents the result of listing transactions.
type TransactionListResult struct {
	Transactions []*entity.TransactionWithCategory
	Total        int64
	Page         int
	Limit        int
	TotalPages   int
}

// TransactionTotals represents aggregated totals for transactions.
type TransactionTotals struct {
	IncomeTotal  decimal.Decimal
	ExpenseTotal decimal.Decimal
	NetTotal     decimal.Decimal
}

// TransactionRepository defines the interface for transaction persistence operations.
type TransactionRepository interface {
	// Create creates a new transaction in the database.
	Create(ctx context.Context, transaction *entity.Transaction) error

	// FindByID retrieves a transaction by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)

	// FindByIDWithCategory retrieves a transaction with its category by ID.
	FindByIDWithCategory(ctx context.Context, id uuid.UUID) (*entity.TransactionWithCategory, error)

	// FindByUser retrieves all transactions for a given user.
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error)

	// FindByFilter retrieves transactions based on filter criteria with pagination.
	FindByFilter(ctx context.Context, filter TransactionFilter, pagination TransactionPagination) (*TransactionListResult, error)

	// GetTotals calculates totals for transactions based on filter criteria.
	GetTotals(ctx context.Context, filter TransactionFilter) (*TransactionTotals, error)

	// Update updates an existing transaction in the database.
	Update(ctx context.Context, transaction *entity.Transaction) error

	// Delete soft-deletes a transaction from the database.
	Delete(ctx context.Context, id uuid.UUID) error

	// BulkDelete soft-deletes multiple transactions by their IDs.
	// Returns the count of deleted transactions.
	BulkDelete(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error)

	// BulkUpdateCategory updates the category for multiple transactions.
	// Returns the count of updated transactions.
	BulkUpdateCategory(ctx context.Context, ids []uuid.UUID, categoryID uuid.UUID, userID uuid.UUID) (int64, error)

	// ExistsByIDAndUser checks if a transaction exists for a given ID and user.
	ExistsByIDAndUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)

	// ExistsAllByIDsAndUser checks if all transactions exist for the given IDs and user.
	ExistsAllByIDsAndUser(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (bool, error)

	// BulkUpdateCategoryByPattern updates category for all uncategorized transactions
	// matching the given pattern for the specified owner.
	BulkUpdateCategoryByPattern(
		ctx context.Context,
		pattern string,
		categoryID uuid.UUID,
		ownerType entity.OwnerType,
		ownerID uuid.UUID,
	) (int, error)

	// Credit card import methods

	// FindPotentialBillPayments finds potential bill payment matches for CC import.
	// It searches for transactions matching "Pagamento de fatura" or similar patterns.
	// Returns transactions within the specified date range that could match CC payments.
	FindPotentialBillPayments(
		ctx context.Context,
		userID uuid.UUID,
		startDate time.Time,
		endDate time.Time,
	) ([]*entity.Transaction, error)

	// GetLinkedTransactions retrieves all CC transactions linked to a bill payment.
	GetLinkedTransactions(ctx context.Context, billPaymentID uuid.UUID) ([]*entity.Transaction, error)

	// BulkCreateCCTransactions creates multiple CC transactions in a single operation.
	// It also updates the bill payment (zeroing amount, setting expanded_at, billing_cycle, etc.).
	BulkCreateCCTransactions(
		ctx context.Context,
		transactions []*entity.Transaction,
		billPaymentID uuid.UUID,
		originalAmount decimal.Decimal,
		billingCycle string,
	) error

	// BulkCreateStandaloneCCTransactions creates CC transactions without linking to a bill payment.
	// Used when importing CC transactions without a matching bill.
	BulkCreateStandaloneCCTransactions(ctx context.Context, transactions []*entity.Transaction) error

	// ExpandBillPayment marks a bill payment as expanded and zeroes its amount.
	ExpandBillPayment(
		ctx context.Context,
		billPaymentID uuid.UUID,
		originalAmount decimal.Decimal,
	) error

	// CollapseExpansion deletes all linked CC transactions and restores the bill payment.
	CollapseExpansion(ctx context.Context, billPaymentID uuid.UUID) error

	// GetCreditCardStatus retrieves the CC status for a specific billing cycle.
	GetCreditCardStatus(
		ctx context.Context,
		userID uuid.UUID,
		billingCycle string,
	) (*CreditCardStatus, error)

	// IsBillExpanded checks if a bill payment has been expanded.
	IsBillExpanded(ctx context.Context, billPaymentID uuid.UUID) (bool, error)

	// FindBillPaymentByID retrieves a bill payment transaction by ID with ownership check.
	FindBillPaymentByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.Transaction, error)

	// FindMostRecentCCBillingCycle finds the most recent billing cycle with CC transactions.
	// Returns empty string if no CC transactions exist.
	FindMostRecentCCBillingCycle(ctx context.Context, userID uuid.UUID) (string, error)

	// GetExpensesByDateRange returns all expense transactions for a user
	// within the specified date range, including category info.
	// Only returns transactions with a category assigned (category_id IS NOT NULL).
	GetExpensesByDateRange(
		ctx context.Context,
		userID uuid.UUID,
		startDate time.Time,
		endDate time.Time,
	) ([]*entity.ExpenseWithCategory, error)

	// CountUncategorizedByUser counts all transactions for a user that have no category assigned.
	// This is used by the AI categorization feature to determine how many transactions need categorization.
	CountUncategorizedByUser(ctx context.Context, userID uuid.UUID) (int, error)
}

// CreditCardStatus represents the status of credit card transactions for a billing cycle.
type CreditCardStatus struct {
	BillingCycle       string
	IsExpanded         bool
	BillPaymentID      *uuid.UUID
	BillPaymentDate    *time.Time
	OriginalAmount     *decimal.Decimal
	CurrentAmount      *decimal.Decimal
	LinkedTransactions []*entity.Transaction
	ExpandedAt         *time.Time
}

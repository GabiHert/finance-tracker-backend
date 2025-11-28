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
}

// Package transaction contains transaction-related use cases.
package transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// BulkDeleteTransactionsInput represents the input for bulk transaction deletion.
type BulkDeleteTransactionsInput struct {
	TransactionIDs []uuid.UUID
	UserID         uuid.UUID
}

// BulkDeleteTransactionsOutput represents the output of bulk transaction deletion.
type BulkDeleteTransactionsOutput struct {
	DeletedCount int64
}

// BulkDeleteTransactionsUseCase handles bulk transaction deletion logic.
type BulkDeleteTransactionsUseCase struct {
	transactionRepo adapter.TransactionRepository
}

// NewBulkDeleteTransactionsUseCase creates a new BulkDeleteTransactionsUseCase instance.
func NewBulkDeleteTransactionsUseCase(transactionRepo adapter.TransactionRepository) *BulkDeleteTransactionsUseCase {
	return &BulkDeleteTransactionsUseCase{
		transactionRepo: transactionRepo,
	}
}

// Execute performs the bulk transaction deletion.
func (uc *BulkDeleteTransactionsUseCase) Execute(ctx context.Context, input BulkDeleteTransactionsInput) (*BulkDeleteTransactionsOutput, error) {
	// Validate that IDs list is not empty
	if len(input.TransactionIDs) == 0 {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeEmptyTransactionIDs,
			"transaction IDs list cannot be empty",
			domainerror.ErrEmptyTransactionIDs,
		)
	}

	// Verify all transactions exist and belong to the user
	allExist, err := uc.transactionRepo.ExistsAllByIDsAndUser(ctx, input.TransactionIDs, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify transactions: %w", err)
	}
	if !allExist {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeTransactionNotFound,
			"one or more transactions not found or not owned by user",
			domainerror.ErrTransactionNotFound,
		)
	}

	// Perform bulk delete (atomic operation)
	deletedCount, err := uc.transactionRepo.BulkDelete(ctx, input.TransactionIDs, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk delete transactions: %w", err)
	}

	return &BulkDeleteTransactionsOutput{
		DeletedCount: deletedCount,
	}, nil
}

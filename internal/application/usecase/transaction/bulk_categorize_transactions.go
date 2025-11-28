// Package transaction contains transaction-related use cases.
package transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// BulkCategorizeTransactionsInput represents the input for bulk transaction categorization.
type BulkCategorizeTransactionsInput struct {
	TransactionIDs []uuid.UUID
	CategoryID     uuid.UUID
	UserID         uuid.UUID
}

// BulkCategorizeTransactionsOutput represents the output of bulk transaction categorization.
type BulkCategorizeTransactionsOutput struct {
	UpdatedCount int64
}

// BulkCategorizeTransactionsUseCase handles bulk transaction categorization logic.
type BulkCategorizeTransactionsUseCase struct {
	transactionRepo adapter.TransactionRepository
	categoryRepo    adapter.CategoryRepository
}

// NewBulkCategorizeTransactionsUseCase creates a new BulkCategorizeTransactionsUseCase instance.
func NewBulkCategorizeTransactionsUseCase(
	transactionRepo adapter.TransactionRepository,
	categoryRepo adapter.CategoryRepository,
) *BulkCategorizeTransactionsUseCase {
	return &BulkCategorizeTransactionsUseCase{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

// Execute performs the bulk transaction categorization.
func (uc *BulkCategorizeTransactionsUseCase) Execute(ctx context.Context, input BulkCategorizeTransactionsInput) (*BulkCategorizeTransactionsOutput, error) {
	// Validate that IDs list is not empty
	if len(input.TransactionIDs) == 0 {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeEmptyTransactionIDs,
			"transaction IDs list cannot be empty",
			domainerror.ErrEmptyTransactionIDs,
		)
	}

	// Validate category exists and belongs to user
	category, err := uc.categoryRepo.FindByID(ctx, input.CategoryID)
	if err != nil {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeTxnCategoryNotFound,
			"category not found",
			domainerror.ErrCategoryNotFoundForTransaction,
		)
	}

	// Verify category ownership
	if category.OwnerType != entity.OwnerTypeUser || category.OwnerID != input.UserID {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeTxnCategoryNotOwned,
			"category does not belong to user",
			domainerror.ErrCategoryNotOwnedByUser,
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

	// Perform bulk category update (atomic operation)
	updatedCount, err := uc.transactionRepo.BulkUpdateCategory(ctx, input.TransactionIDs, input.CategoryID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk categorize transactions: %w", err)
	}

	return &BulkCategorizeTransactionsOutput{
		UpdatedCount: updatedCount,
	}, nil
}

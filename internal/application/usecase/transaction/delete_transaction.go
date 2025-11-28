// Package transaction contains transaction-related use cases.
package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// DeleteTransactionInput represents the input for transaction deletion.
type DeleteTransactionInput struct {
	TransactionID uuid.UUID
	UserID        uuid.UUID
}

// DeleteTransactionOutput represents the output of transaction deletion.
type DeleteTransactionOutput struct {
	Success bool
}

// DeleteTransactionUseCase handles transaction deletion logic.
type DeleteTransactionUseCase struct {
	transactionRepo adapter.TransactionRepository
}

// NewDeleteTransactionUseCase creates a new DeleteTransactionUseCase instance.
func NewDeleteTransactionUseCase(transactionRepo adapter.TransactionRepository) *DeleteTransactionUseCase {
	return &DeleteTransactionUseCase{
		transactionRepo: transactionRepo,
	}
}

// Execute performs the transaction deletion.
func (uc *DeleteTransactionUseCase) Execute(ctx context.Context, input DeleteTransactionInput) (*DeleteTransactionOutput, error) {
	// Find the existing transaction
	transaction, err := uc.transactionRepo.FindByID(ctx, input.TransactionID)
	if err != nil {
		if errors.Is(err, domainerror.ErrTransactionNotFound) {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeTransactionNotFound,
				"transaction not found",
				domainerror.ErrTransactionNotFound,
			)
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}

	// Check if user is authorized to delete this transaction
	if transaction.UserID != input.UserID {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeNotAuthorizedTransaction,
			"not authorized to delete this transaction",
			domainerror.ErrNotAuthorizedToModifyTransaction,
		)
	}

	// Delete the transaction (soft delete)
	if err := uc.transactionRepo.Delete(ctx, input.TransactionID); err != nil {
		return nil, fmt.Errorf("failed to delete transaction: %w", err)
	}

	return &DeleteTransactionOutput{
		Success: true,
	}, nil
}

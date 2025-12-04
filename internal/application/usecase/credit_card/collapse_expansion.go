// Package creditcard contains credit card import-related use cases.
package creditcard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// CollapseExpansionInput represents the input for collapsing CC expansion.
type CollapseExpansionInput struct {
	UserID        uuid.UUID
	BillPaymentID uuid.UUID
}

// CollapseExpansionOutput represents the output of collapse operation.
type CollapseExpansionOutput struct {
	BillPaymentID       uuid.UUID
	RestoredAmount      decimal.Decimal
	DeletedTransactions int
	CollapsedAt         time.Time
}

// CollapseExpansionUseCase handles the CC expansion collapse logic.
type CollapseExpansionUseCase struct {
	transactionRepo adapter.TransactionRepository
}

// NewCollapseExpansionUseCase creates a new CollapseExpansionUseCase instance.
func NewCollapseExpansionUseCase(transactionRepo adapter.TransactionRepository) *CollapseExpansionUseCase {
	return &CollapseExpansionUseCase{
		transactionRepo: transactionRepo,
	}
}

// Execute performs the CC expansion collapse operation.
func (uc *CollapseExpansionUseCase) Execute(ctx context.Context, input CollapseExpansionInput) (*CollapseExpansionOutput, error) {
	// Verify bill payment exists and belongs to user
	billPayment, err := uc.transactionRepo.FindBillPaymentByID(ctx, input.BillPaymentID, input.UserID)
	if err != nil {
		if err == domainerror.ErrBillPaymentNotFound {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeBillPaymentNotFound,
				"bill payment transaction not found",
				domainerror.ErrBillPaymentNotFound,
			)
		}
		return nil, err
	}

	// Check if bill is actually expanded
	isExpanded, err := uc.transactionRepo.IsBillExpanded(ctx, input.BillPaymentID)
	if err != nil {
		return nil, err
	}
	if !isExpanded {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeBillNotExpanded,
			"bill is not expanded",
			domainerror.ErrBillNotExpanded,
		)
	}

	// Get linked transactions count before deletion
	linkedTransactions, err := uc.transactionRepo.GetLinkedTransactions(ctx, input.BillPaymentID)
	if err != nil {
		return nil, err
	}

	// Get the original amount that will be restored
	restoredAmount := decimal.Zero
	if billPayment.OriginalAmount != nil {
		restoredAmount = *billPayment.OriginalAmount
	}

	// Collapse the expansion (delete linked transactions and restore bill)
	if err := uc.transactionRepo.CollapseExpansion(ctx, input.BillPaymentID); err != nil {
		return nil, err
	}

	return &CollapseExpansionOutput{
		BillPaymentID:       input.BillPaymentID,
		RestoredAmount:      restoredAmount,
		DeletedTransactions: len(linkedTransactions),
		CollapsedAt:         time.Now().UTC(),
	}, nil
}

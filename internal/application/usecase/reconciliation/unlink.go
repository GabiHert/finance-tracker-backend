// Package reconciliation contains credit card reconciliation use cases.
package reconciliation

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// UnlinkInput represents the input for unlinking a billing cycle from its bill.
type UnlinkInput struct {
	UserID       uuid.UUID
	BillingCycle string
}

// UnlinkOutput represents the result of unlinking.
type UnlinkOutput struct {
	BillingCycle string
	Success      bool
}

// UnlinkUseCase handles unlinking CC transactions from a bill.
type UnlinkUseCase struct {
	reconciliationRepo adapter.ReconciliationRepository
}

// NewUnlinkUseCase creates a new UnlinkUseCase instance.
func NewUnlinkUseCase(reconciliationRepo adapter.ReconciliationRepository) *UnlinkUseCase {
	return &UnlinkUseCase{
		reconciliationRepo: reconciliationRepo,
	}
}

// Execute performs the unlinking operation.
func (uc *UnlinkUseCase) Execute(ctx context.Context, input UnlinkInput) (*UnlinkOutput, error) {
	// Validate billing cycle format
	if !billingCycleRegex.MatchString(input.BillingCycle) {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeInvalidBillingCycle,
			"billing cycle must be in YYYY-MM format",
			domainerror.ErrInvalidBillingCycle,
		)
	}

	// Check if the billing cycle is linked
	isLinked, _, err := uc.reconciliationRepo.IsCycleLinked(ctx, input.UserID, input.BillingCycle)
	if err != nil {
		return nil, err
	}
	if !isLinked {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodePendingNotFound,
			"billing cycle is not linked to any bill",
			domainerror.ErrPendingNotFound,
		)
	}

	// Perform the unlinking
	err = uc.reconciliationRepo.UnlinkCCTransactionsFromBill(ctx, input.UserID, input.BillingCycle)
	if err != nil {
		return nil, err
	}

	return &UnlinkOutput{
		BillingCycle: input.BillingCycle,
		Success:      true,
	}, nil
}

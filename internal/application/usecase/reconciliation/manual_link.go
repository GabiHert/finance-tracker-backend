// Package reconciliation contains credit card reconciliation use cases.
package reconciliation

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/domain/valueobject"
)

// BillingCyclePattern is the regex pattern for valid billing cycle format (YYYY-MM).
const BillingCyclePattern = `^\d{4}-(0[1-9]|1[0-2])$`

var billingCycleRegex = regexp.MustCompile(BillingCyclePattern)

// ManualLinkInput represents the input for manually linking a billing cycle to a bill.
type ManualLinkInput struct {
	UserID       uuid.UUID
	BillingCycle string
	BillID       uuid.UUID
	Force        bool // If true, allow linking even with large mismatch
}

// ManualLinkOutput represents the result of manual linking.
type ManualLinkOutput struct {
	BillingCycle       string
	BillID             uuid.UUID
	TransactionsLinked int
	AmountDifference   decimal.Decimal
	HasMismatch        bool
}

// ManualLinkUseCase handles manually linking CC transactions to a bill.
type ManualLinkUseCase struct {
	reconciliationRepo adapter.ReconciliationRepository
	config             valueobject.MatchingConfig
}

// NewManualLinkUseCase creates a new ManualLinkUseCase instance.
func NewManualLinkUseCase(reconciliationRepo adapter.ReconciliationRepository) *ManualLinkUseCase {
	return &ManualLinkUseCase{
		reconciliationRepo: reconciliationRepo,
		config:             valueobject.DefaultMatchingConfig(),
	}
}

// Execute performs the manual linking operation.
func (uc *ManualLinkUseCase) Execute(ctx context.Context, input ManualLinkInput) (*ManualLinkOutput, error) {
	// Validate billing cycle format
	if !billingCycleRegex.MatchString(input.BillingCycle) {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeInvalidBillingCycle,
			"billing cycle must be in YYYY-MM format",
			domainerror.ErrInvalidBillingCycle,
		)
	}

	// Check if the billing cycle has pending CC transactions
	ccTotal, transactionCount, err := uc.reconciliationRepo.GetCCTotalByBillingCycle(ctx, input.UserID, input.BillingCycle)
	if err != nil {
		return nil, err
	}
	if transactionCount == 0 {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodePendingNotFound,
			"no pending CC transactions for billing cycle",
			domainerror.ErrPendingNotFound,
		)
	}

	// Check if the billing cycle is already linked
	isLinked, existingBillID, err := uc.reconciliationRepo.IsCycleLinked(ctx, input.UserID, input.BillingCycle)
	if err != nil {
		return nil, err
	}
	if isLinked && existingBillID != nil {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeCycleAlreadyLinked,
			"billing cycle already has linked bill",
			domainerror.ErrCycleAlreadyLinked,
		)
	}

	// Verify the bill payment exists and belongs to user
	billData, err := uc.reconciliationRepo.GetBillPaymentByID(ctx, input.BillID, input.UserID)
	if err != nil {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeBillPaymentNotFound,
			"bill payment not found",
			domainerror.ErrBillPaymentNotFound,
		)
	}

	// Check if the bill is already linked to another cycle
	isBillLinked, err := uc.reconciliationRepo.IsBillLinked(ctx, input.BillID)
	if err != nil {
		return nil, err
	}
	if isBillLinked {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeBillAlreadyExpanded,
			"bill is already linked to another cycle",
			domainerror.ErrBillAlreadyExpanded,
		)
	}

	// Calculate amount difference
	billAmount := decimal.NewFromInt(billData.Amount)
	diff := ccTotal.Sub(billAmount)

	// Check if within tolerance (unless force is true)
	hasMismatch := !uc.config.IsWithinTolerance(ccTotal, billAmount)
	if hasMismatch && !input.Force {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeAmountMismatch,
			"amount difference exceeds tolerance, use force to override",
			domainerror.ErrAmountMismatch,
		)
	}

	// Perform the linking
	linkedCount, err := uc.reconciliationRepo.LinkCCTransactionsToBill(
		ctx, input.UserID, input.BillingCycle, input.BillID, billAmount,
	)
	if err != nil {
		return nil, err
	}

	return &ManualLinkOutput{
		BillingCycle:       input.BillingCycle,
		BillID:             input.BillID,
		TransactionsLinked: linkedCount,
		AmountDifference:   diff,
		HasMismatch:        hasMismatch,
	}, nil
}

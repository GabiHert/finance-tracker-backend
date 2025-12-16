// Package reconciliation contains credit card reconciliation use cases.
package reconciliation

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/valueobject"
)

// GetLinkedInput represents the input for getting linked reconciliations.
type GetLinkedInput struct {
	UserID uuid.UUID
	Limit  int
	Offset int
}

// LinkedCycleOutput represents a linked billing cycle.
type LinkedCycleOutput struct {
	BillingCycle     string
	DisplayName      string
	TransactionCount int
	TotalAmount      decimal.Decimal
	Bill             LinkedBillOutput
	AmountDifference decimal.Decimal
	HasMismatch      bool
}

// LinkedBillOutput contains information about the linked bill payment.
type LinkedBillOutput struct {
	ID             uuid.UUID
	Date           time.Time
	Description    string
	OriginalAmount decimal.Decimal
	CategoryName   *string
}

// GetLinkedOutput represents the output for getting linked reconciliations.
type GetLinkedOutput struct {
	LinkedCycles []LinkedCycleOutput
	Summary      ReconciliationSummaryOutput
}

// GetLinkedUseCase handles getting linked reconciliations.
type GetLinkedUseCase struct {
	reconciliationRepo adapter.ReconciliationRepository
	config             valueobject.MatchingConfig
}

// NewGetLinkedUseCase creates a new GetLinkedUseCase instance.
func NewGetLinkedUseCase(reconciliationRepo adapter.ReconciliationRepository) *GetLinkedUseCase {
	return &GetLinkedUseCase{
		reconciliationRepo: reconciliationRepo,
		config:             valueobject.DefaultMatchingConfig(),
	}
}

// Execute retrieves linked billing cycles.
func (uc *GetLinkedUseCase) Execute(ctx context.Context, input GetLinkedInput) (*GetLinkedOutput, error) {
	// Set defaults
	limit := input.Limit
	if limit <= 0 {
		limit = 12 // Default to 12 months
	}

	// Get linked billing cycles
	linkedCycles, err := uc.reconciliationRepo.GetLinkedBillingCycles(ctx, input.UserID, limit, input.Offset)
	if err != nil {
		return nil, err
	}

	// Get summary
	summary, err := uc.reconciliationRepo.GetReconciliationSummary(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Build output
	outputCycles := make([]LinkedCycleOutput, 0, len(linkedCycles))

	for _, cycle := range linkedCycles {
		ccTotal := decimal.NewFromInt(cycle.TotalAmount)
		billAmount := decimal.NewFromInt(cycle.BillAmount)
		diff := ccTotal.Sub(billAmount)

		// Check if there's a mismatch (beyond tolerance)
		hasMismatch := !uc.config.IsWithinTolerance(ccTotal, billAmount)

		outputCycles = append(outputCycles, LinkedCycleOutput{
			BillingCycle:     cycle.BillingCycle,
			DisplayName:      valueobject.FormatBillingCycleDisplay(cycle.BillingCycle),
			TransactionCount: cycle.TransactionCount,
			TotalAmount:      ccTotal,
			Bill: LinkedBillOutput{
				ID:             cycle.BillID,
				Date:           cycle.BillDate,
				Description:    cycle.BillDescription,
				OriginalAmount: billAmount,
				CategoryName:   cycle.CategoryName,
			},
			AmountDifference: diff,
			HasMismatch:      hasMismatch,
		})
	}

	return &GetLinkedOutput{
		LinkedCycles: outputCycles,
		Summary: ReconciliationSummaryOutput{
			TotalPending:  summary.TotalPending,
			TotalLinked:   summary.TotalLinked,
			MonthsCovered: summary.MonthsCovered,
		},
	}, nil
}

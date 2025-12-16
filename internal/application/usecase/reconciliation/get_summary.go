// Package reconciliation contains credit card reconciliation use cases.
package reconciliation

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
)

// GetSummaryInput represents the input for getting reconciliation summary.
type GetSummaryInput struct {
	UserID uuid.UUID
}

// GetSummaryOutput represents the output for getting reconciliation summary.
type GetSummaryOutput struct {
	TotalPending  int
	TotalLinked   int
	MonthsCovered int
}

// GetSummaryUseCase handles getting reconciliation summary.
type GetSummaryUseCase struct {
	reconciliationRepo adapter.ReconciliationRepository
}

// NewGetSummaryUseCase creates a new GetSummaryUseCase instance.
func NewGetSummaryUseCase(reconciliationRepo adapter.ReconciliationRepository) *GetSummaryUseCase {
	return &GetSummaryUseCase{
		reconciliationRepo: reconciliationRepo,
	}
}

// Execute retrieves reconciliation summary statistics.
func (uc *GetSummaryUseCase) Execute(ctx context.Context, input GetSummaryInput) (*GetSummaryOutput, error) {
	summary, err := uc.reconciliationRepo.GetReconciliationSummary(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetSummaryOutput{
		TotalPending:  summary.TotalPending,
		TotalLinked:   summary.TotalLinked,
		MonthsCovered: summary.MonthsCovered,
	}, nil
}

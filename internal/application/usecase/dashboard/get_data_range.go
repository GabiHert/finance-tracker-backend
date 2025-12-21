// Package dashboard contains dashboard-related use cases.
package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GetDataRangeInput represents the input for getting data range.
type GetDataRangeInput struct {
	UserID uuid.UUID
}

// GetDataRangeOutput represents the output of getting data range.
type GetDataRangeOutput struct {
	OldestDate        *time.Time `json:"oldest_date"`
	NewestDate        *time.Time `json:"newest_date"`
	TotalTransactions int        `json:"total_transactions"`
	HasData           bool       `json:"has_data"`
}

// GetDataRangeUseCase handles getting the date range of user's transactions.
type GetDataRangeUseCase struct {
	dashboardRepo DashboardRepository
}

// NewGetDataRangeUseCase creates a new GetDataRangeUseCase instance.
func NewGetDataRangeUseCase(dashboardRepo DashboardRepository) *GetDataRangeUseCase {
	return &GetDataRangeUseCase{
		dashboardRepo: dashboardRepo,
	}
}

// Execute retrieves the date range of user's transactions.
func (uc *GetDataRangeUseCase) Execute(
	ctx context.Context,
	input GetDataRangeInput,
) (*GetDataRangeOutput, error) {
	dateRange, err := uc.dashboardRepo.GetDateRange(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get date range: %w", err)
	}

	hasData := dateRange.OldestDate != nil && dateRange.NewestDate != nil

	return &GetDataRangeOutput{
		OldestDate:        dateRange.OldestDate,
		NewestDate:        dateRange.NewestDate,
		TotalTransactions: dateRange.TotalTransactions,
		HasData:           hasData,
	}, nil
}

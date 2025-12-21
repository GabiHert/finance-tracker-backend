// Package dashboard contains dashboard-related use cases.
package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// GetTrendsInput represents the input for getting trends.
type GetTrendsInput struct {
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
	Granularity Granularity
}

// TrendPoint represents a single trend data point.
type TrendPoint struct {
	Date             time.Time       `json:"date"`
	PeriodLabel      string          `json:"period_label"`
	Income           decimal.Decimal `json:"income"`
	Expenses         decimal.Decimal `json:"expenses"`
	Balance          decimal.Decimal `json:"balance"`
	TransactionCount int             `json:"transaction_count"`
}

// GetTrendsOutput represents the output of getting trends.
type GetTrendsOutput struct {
	Period TrendsPeriod `json:"period"`
	Trends []TrendPoint `json:"trends"`
}

// TrendsPeriod represents the period information for trends.
type TrendsPeriod struct {
	StartDate   time.Time   `json:"start_date"`
	EndDate     time.Time   `json:"end_date"`
	Granularity Granularity `json:"granularity"`
}

// GetTrendsUseCase handles getting income/expense trends.
type GetTrendsUseCase struct {
	dashboardRepo DashboardRepository
}

// NewGetTrendsUseCase creates a new GetTrendsUseCase instance.
func NewGetTrendsUseCase(dashboardRepo DashboardRepository) *GetTrendsUseCase {
	return &GetTrendsUseCase{
		dashboardRepo: dashboardRepo,
	}
}

// Execute retrieves income/expense trends for the given period and granularity.
func (uc *GetTrendsUseCase) Execute(
	ctx context.Context,
	input GetTrendsInput,
) (*GetTrendsOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Get raw trend data from repository
	rawTrends, err := uc.dashboardRepo.GetAggregatedTrends(
		ctx,
		input.UserID,
		input.StartDate,
		input.EndDate,
		input.Granularity,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get trends: %w", err)
	}

	// Create a map for quick lookup of raw data by period key
	rawDataMap := make(map[string]RawTrendData)
	for _, rd := range rawTrends {
		key := GetPeriodKeyForDate(rd.PeriodStart, input.Granularity)
		rawDataMap[key] = rd
	}

	// Generate all periods in the range to ensure no gaps
	periods := GeneratePeriodSeries(input.StartDate, input.EndDate, input.Granularity)

	// Build trend points with zero values for empty periods
	trends := make([]TrendPoint, 0, len(periods))
	for _, period := range periods {
		key := period.Date.Format("2006-01-02")

		// Check if we have data for this period
		if rawData, ok := rawDataMap[key]; ok {
			balance := rawData.Income.Sub(rawData.Expenses)
			trends = append(trends, TrendPoint{
				Date:             period.Date,
				PeriodLabel:      period.PeriodLabel,
				Income:           rawData.Income,
				Expenses:         rawData.Expenses,
				Balance:          balance,
				TransactionCount: rawData.TransactionCount,
			})
		} else {
			// Empty period - include with zero values
			trends = append(trends, TrendPoint{
				Date:             period.Date,
				PeriodLabel:      period.PeriodLabel,
				Income:           decimal.Zero,
				Expenses:         decimal.Zero,
				Balance:          decimal.Zero,
				TransactionCount: 0,
			})
		}
	}

	return &GetTrendsOutput{
		Period: TrendsPeriod{
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			Granularity: input.Granularity,
		},
		Trends: trends,
	}, nil
}

// validateInput validates the input parameters.
func (uc *GetTrendsUseCase) validateInput(input GetTrendsInput) error {
	if input.StartDate.IsZero() {
		return domainerror.NewDashboardError(
			domainerror.ErrCodeMissingStartDate,
			"start_date is required",
			domainerror.ErrMissingStartDate,
		)
	}

	if input.EndDate.IsZero() {
		return domainerror.NewDashboardError(
			domainerror.ErrCodeMissingEndDate,
			"end_date is required",
			domainerror.ErrMissingEndDate,
		)
	}

	if input.EndDate.Before(input.StartDate) {
		return domainerror.NewDashboardError(
			domainerror.ErrCodeInvalidDateRange,
			"end_date must be after start_date",
			domainerror.ErrInvalidDateRange,
		)
	}

	if input.Granularity != GranularityWeekly &&
		input.Granularity != GranularityMonthly &&
		input.Granularity != GranularityQuarterly {
		return domainerror.NewDashboardError(
			domainerror.ErrCodeInvalidGranularity,
			"granularity must be: weekly, monthly, or quarterly",
			domainerror.ErrInvalidGranularity,
		)
	}

	return nil
}

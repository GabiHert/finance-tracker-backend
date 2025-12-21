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

// UncategorizedID is a constant string used to represent uncategorized transactions.
const UncategorizedID = "uncategorized"

// UncategorizedName is the default name for uncategorized transactions (Portuguese).
const UncategorizedName = "Sem categoria"

// UncategorizedColor is the default color for uncategorized transactions.
const UncategorizedColor = "#6B7280"

// UncategorizedIcon is the default icon for uncategorized transactions.
const UncategorizedIcon = "question-mark"

// GetCategoryBreakdownInput represents the input for getting category breakdown.
type GetCategoryBreakdownInput struct {
	UserID    uuid.UUID
	StartDate time.Time
	EndDate   time.Time
}

// CategoryBreakdownItem represents a single category in the breakdown.
type CategoryBreakdownItem struct {
	CategoryID       string          `json:"category_id"`
	CategoryName     string          `json:"category_name"`
	CategoryColor    string          `json:"category_color"`
	CategoryIcon     string          `json:"category_icon"`
	Amount           decimal.Decimal `json:"amount"`
	Percentage       float64         `json:"percentage"`
	TransactionCount int             `json:"transaction_count"`
}

// GetCategoryBreakdownOutput represents the output of getting category breakdown.
type GetCategoryBreakdownOutput struct {
	Period        BreakdownPeriod         `json:"period"`
	TotalExpenses decimal.Decimal         `json:"total_expenses"`
	Categories    []CategoryBreakdownItem `json:"categories"`
}

// BreakdownPeriod represents the period information for category breakdown.
type BreakdownPeriod struct {
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	PeriodLabel string    `json:"period_label"`
}

// GetCategoryBreakdownUseCase handles getting spending breakdown by category.
type GetCategoryBreakdownUseCase struct {
	dashboardRepo DashboardRepository
}

// NewGetCategoryBreakdownUseCase creates a new GetCategoryBreakdownUseCase instance.
func NewGetCategoryBreakdownUseCase(dashboardRepo DashboardRepository) *GetCategoryBreakdownUseCase {
	return &GetCategoryBreakdownUseCase{
		dashboardRepo: dashboardRepo,
	}
}

// Execute retrieves spending breakdown by category for the given period.
func (uc *GetCategoryBreakdownUseCase) Execute(
	ctx context.Context,
	input GetCategoryBreakdownInput,
) (*GetCategoryBreakdownOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Get category breakdown from repository
	rawBreakdown, totalExpenses, err := uc.dashboardRepo.GetCategoryBreakdown(
		ctx,
		input.UserID,
		input.StartDate,
		input.EndDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get category breakdown: %w", err)
	}

	// Convert raw data to output format
	categories := make([]CategoryBreakdownItem, 0, len(rawBreakdown))
	for _, raw := range rawBreakdown {
		var percentage float64
		if !totalExpenses.IsZero() {
			pct := raw.Amount.Mul(decimal.NewFromInt(100)).Div(totalExpenses)
			percentage, _ = pct.Round(2).Float64()
		}

		item := CategoryBreakdownItem{
			Amount:           raw.Amount,
			Percentage:       percentage,
			TransactionCount: raw.TransactionCount,
		}

		// Handle uncategorized transactions
		if raw.CategoryID == nil {
			item.CategoryID = UncategorizedID
			item.CategoryName = UncategorizedName
			item.CategoryColor = UncategorizedColor
			item.CategoryIcon = UncategorizedIcon
		} else {
			item.CategoryID = raw.CategoryID.String()
			if raw.CategoryName != nil {
				item.CategoryName = *raw.CategoryName
			}
			if raw.CategoryColor != nil {
				item.CategoryColor = *raw.CategoryColor
			} else {
				item.CategoryColor = UncategorizedColor
			}
			if raw.CategoryIcon != nil {
				item.CategoryIcon = *raw.CategoryIcon
			} else {
				item.CategoryIcon = UncategorizedIcon
			}
		}

		categories = append(categories, item)
	}

	// Generate period label based on the date range
	periodLabel := uc.generatePeriodLabel(input.StartDate, input.EndDate)

	return &GetCategoryBreakdownOutput{
		Period: BreakdownPeriod{
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			PeriodLabel: periodLabel,
		},
		TotalExpenses: totalExpenses,
		Categories:    categories,
	}, nil
}

// validateInput validates the input parameters.
func (uc *GetCategoryBreakdownUseCase) validateInput(input GetCategoryBreakdownInput) error {
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

	return nil
}

// generatePeriodLabel generates a human-readable label for the period.
func (uc *GetCategoryBreakdownUseCase) generatePeriodLabel(startDate, endDate time.Time) string {
	// If the period spans a single month, use monthly format
	if startDate.Year() == endDate.Year() && startDate.Month() == endDate.Month() {
		return GeneratePeriodLabel(startDate, GranularityMonthly)
	}

	// If the period spans a single quarter
	startQuarter := (int(startDate.Month())-1)/3 + 1
	endQuarter := (int(endDate.Month())-1)/3 + 1
	if startDate.Year() == endDate.Year() && startQuarter == endQuarter {
		return GeneratePeriodLabel(startDate, GranularityQuarterly)
	}

	// Otherwise, show the date range
	return fmt.Sprintf("%s - %s",
		GeneratePeriodLabel(startDate, GranularityMonthly),
		GeneratePeriodLabel(endDate, GranularityMonthly),
	)
}

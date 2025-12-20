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

// DefaultTransactionLimit is the default limit for transactions per page.
const DefaultTransactionLimit = 50

// MaxTransactionLimit is the maximum allowed limit for transactions per page.
const MaxTransactionLimit = 100

// GetPeriodTransactionsInput represents the input for getting period transactions.
type GetPeriodTransactionsInput struct {
	UserID     uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
	CategoryID *uuid.UUID // Optional filter
	Limit      int
	Offset     int
}

// PeriodTransactionItem represents a transaction within a period.
type PeriodTransactionItem struct {
	ID            string           `json:"id"`
	Description   string           `json:"description"`
	Amount        decimal.Decimal  `json:"amount"`
	Date          time.Time        `json:"date"`
	CategoryID    *string          `json:"category_id,omitempty"`
	CategoryName  *string          `json:"category_name,omitempty"`
	CategoryColor *string          `json:"category_color,omitempty"`
	CategoryIcon  *string          `json:"category_icon,omitempty"`
}

// TransactionSummary represents summary totals for transactions.
type TransactionSummary struct {
	TotalIncome      decimal.Decimal `json:"total_income"`
	TotalExpenses    decimal.Decimal `json:"total_expenses"`
	Balance          decimal.Decimal `json:"balance"`
	TransactionCount int             `json:"transaction_count"`
}

// TransactionPagination represents pagination information.
type TransactionPagination struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

// GetPeriodTransactionsOutput represents the output of getting period transactions.
type GetPeriodTransactionsOutput struct {
	Period       TransactionsPeriod      `json:"period"`
	Summary      TransactionSummary      `json:"summary"`
	Transactions []PeriodTransactionItem `json:"transactions"`
	Pagination   TransactionPagination   `json:"pagination"`
}

// TransactionsPeriod represents the period information for transactions.
type TransactionsPeriod struct {
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	PeriodLabel string    `json:"period_label"`
}

// GetPeriodTransactionsUseCase handles getting transactions for a specific period.
type GetPeriodTransactionsUseCase struct {
	dashboardRepo DashboardRepository
}

// NewGetPeriodTransactionsUseCase creates a new GetPeriodTransactionsUseCase instance.
func NewGetPeriodTransactionsUseCase(dashboardRepo DashboardRepository) *GetPeriodTransactionsUseCase {
	return &GetPeriodTransactionsUseCase{
		dashboardRepo: dashboardRepo,
	}
}

// Execute retrieves transactions for the given period with optional category filter.
func (uc *GetPeriodTransactionsUseCase) Execute(
	ctx context.Context,
	input GetPeriodTransactionsInput,
) (*GetPeriodTransactionsOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Apply default/max limits
	limit := input.Limit
	if limit <= 0 {
		limit = DefaultTransactionLimit
	}
	if limit > MaxTransactionLimit {
		limit = MaxTransactionLimit
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	// Get period summary
	summary, err := uc.dashboardRepo.GetPeriodSummary(
		ctx,
		input.UserID,
		input.StartDate,
		input.EndDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get period summary: %w", err)
	}

	// Get transactions
	transactions, total, err := uc.dashboardRepo.GetTransactionsByPeriod(
		ctx,
		input.UserID,
		input.StartDate,
		input.EndDate,
		input.CategoryID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Convert to output format
	transactionItems := make([]PeriodTransactionItem, 0, len(transactions))
	for _, t := range transactions {
		item := PeriodTransactionItem{
			ID:          t.ID.String(),
			Description: t.Description,
			Amount:      t.Amount,
			Date:        t.Date,
		}

		if t.CategoryID != nil {
			catID := t.CategoryID.String()
			item.CategoryID = &catID
		}
		item.CategoryName = t.CategoryName
		item.CategoryColor = t.CategoryColor
		item.CategoryIcon = t.CategoryIcon

		transactionItems = append(transactionItems, item)
	}

	// Generate period label
	periodLabel := uc.generatePeriodLabel(input.StartDate, input.EndDate)

	return &GetPeriodTransactionsOutput{
		Period: TransactionsPeriod{
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			PeriodLabel: periodLabel,
		},
		Summary: TransactionSummary{
			TotalIncome:      summary.TotalIncome,
			TotalExpenses:    summary.TotalExpenses,
			Balance:          summary.Balance,
			TransactionCount: summary.TransactionCount,
		},
		Transactions: transactionItems,
		Pagination: TransactionPagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+len(transactionItems) < total,
		},
	}, nil
}

// validateInput validates the input parameters.
func (uc *GetPeriodTransactionsUseCase) validateInput(input GetPeriodTransactionsInput) error {
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
func (uc *GetPeriodTransactionsUseCase) generatePeriodLabel(startDate, endDate time.Time) string {
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

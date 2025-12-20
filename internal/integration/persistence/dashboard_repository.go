// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/usecase/dashboard"
)

// dashboardRepository implements the dashboard.DashboardRepository interface.
type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates a new dashboard repository instance.
func NewDashboardRepository(db *gorm.DB) dashboard.DashboardRepository {
	return &dashboardRepository{
		db: db,
	}
}

// GetDateRange returns the date range of user's transactions.
func (r *dashboardRepository) GetDateRange(
	ctx context.Context,
	userID uuid.UUID,
) (*dashboard.DateRange, error) {
	var result struct {
		OldestDate *time.Time `gorm:"column:oldest_date"`
		NewestDate *time.Time `gorm:"column:newest_date"`
		Total      int        `gorm:"column:total"`
	}

	err := r.db.WithContext(ctx).
		Table("transactions").
		Select("MIN(date) as oldest_date, MAX(date) as newest_date, COUNT(*) as total").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get date range: %w", err)
	}

	return &dashboard.DateRange{
		OldestDate:        result.OldestDate,
		NewestDate:        result.NewestDate,
		TotalTransactions: result.Total,
	}, nil
}

// GetAggregatedTrends returns income/expense trends aggregated by granularity.
func (r *dashboardRepository) GetAggregatedTrends(
	ctx context.Context,
	userID uuid.UUID,
	startDate, endDate time.Time,
	granularity dashboard.Granularity,
) ([]dashboard.RawTrendData, error) {
	// Determine the date_trunc interval based on granularity
	var truncInterval string
	switch granularity {
	case dashboard.GranularityWeekly:
		truncInterval = "week"
	case dashboard.GranularityMonthly:
		truncInterval = "month"
	case dashboard.GranularityQuarterly:
		truncInterval = "quarter"
	default:
		truncInterval = "month"
	}

	var results []struct {
		PeriodStart      time.Time       `gorm:"column:period_start"`
		Income           decimal.Decimal `gorm:"column:income"`
		Expenses         decimal.Decimal `gorm:"column:expenses"`
		TransactionCount int             `gorm:"column:transaction_count"`
	}

	query := fmt.Sprintf(`
		SELECT
			date_trunc('%s', date)::date as period_start,
			SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) as income,
			SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END) as expenses,
			COUNT(*) as transaction_count
		FROM transactions
		WHERE user_id = ?
			AND date >= ?
			AND date <= ?
			AND deleted_at IS NULL
		GROUP BY date_trunc('%s', date)
		ORDER BY period_start
	`, truncInterval, truncInterval)

	err := r.db.WithContext(ctx).
		Raw(query, userID, startDate, endDate).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated trends: %w", err)
	}

	trends := make([]dashboard.RawTrendData, len(results))
	for i, res := range results {
		trends[i] = dashboard.RawTrendData{
			PeriodStart:      res.PeriodStart,
			Income:           res.Income,
			Expenses:         res.Expenses,
			TransactionCount: res.TransactionCount,
		}
	}

	return trends, nil
}

// GetCategoryBreakdown returns spending breakdown by category for a period.
func (r *dashboardRepository) GetCategoryBreakdown(
	ctx context.Context,
	userID uuid.UUID,
	startDate, endDate time.Time,
) ([]dashboard.RawCategoryBreakdown, decimal.Decimal, error) {
	var results []struct {
		CategoryID       *uuid.UUID      `gorm:"column:category_id"`
		CategoryName     *string         `gorm:"column:category_name"`
		CategoryColor    *string         `gorm:"column:category_color"`
		CategoryIcon     *string         `gorm:"column:category_icon"`
		Amount           decimal.Decimal `gorm:"column:amount"`
		TransactionCount int             `gorm:"column:transaction_count"`
	}

	query := `
		SELECT
			t.category_id,
			c.name as category_name,
			c.color as category_color,
			c.icon as category_icon,
			SUM(ABS(t.amount)) as amount,
			COUNT(*) as transaction_count
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id AND c.deleted_at IS NULL
		WHERE t.user_id = ?
			AND t.date >= ?
			AND t.date <= ?
			AND t.amount < 0
			AND t.deleted_at IS NULL
		GROUP BY t.category_id, c.name, c.color, c.icon
		ORDER BY amount DESC
	`

	err := r.db.WithContext(ctx).
		Raw(query, userID, startDate, endDate).
		Scan(&results).Error

	if err != nil {
		return nil, decimal.Zero, fmt.Errorf("failed to get category breakdown: %w", err)
	}

	// Calculate total expenses
	totalExpenses := decimal.Zero
	breakdown := make([]dashboard.RawCategoryBreakdown, len(results))
	for i, res := range results {
		totalExpenses = totalExpenses.Add(res.Amount)
		breakdown[i] = dashboard.RawCategoryBreakdown{
			CategoryID:       res.CategoryID,
			CategoryName:     res.CategoryName,
			CategoryColor:    res.CategoryColor,
			CategoryIcon:     res.CategoryIcon,
			Amount:           res.Amount,
			TransactionCount: res.TransactionCount,
		}
	}

	return breakdown, totalExpenses, nil
}

// GetTransactionsByPeriod returns transactions for a specific period.
func (r *dashboardRepository) GetTransactionsByPeriod(
	ctx context.Context,
	userID uuid.UUID,
	startDate, endDate time.Time,
	categoryID *uuid.UUID,
	limit, offset int,
) ([]dashboard.PeriodTransaction, int, error) {
	var results []struct {
		ID            uuid.UUID       `gorm:"column:id"`
		Description   string          `gorm:"column:description"`
		Amount        decimal.Decimal `gorm:"column:amount"`
		Date          time.Time       `gorm:"column:date"`
		CategoryID    *uuid.UUID      `gorm:"column:category_id"`
		CategoryName  *string         `gorm:"column:category_name"`
		CategoryColor *string         `gorm:"column:category_color"`
		CategoryIcon  *string         `gorm:"column:category_icon"`
	}

	// Build base query
	baseQuery := r.db.WithContext(ctx).
		Table("transactions t").
		Select(`
			t.id,
			t.description,
			t.amount,
			t.date,
			t.category_id,
			c.name as category_name,
			c.color as category_color,
			c.icon as category_icon
		`).
		Joins("LEFT JOIN categories c ON t.category_id = c.id AND c.deleted_at IS NULL").
		Where("t.user_id = ?", userID).
		Where("t.date >= ?", startDate).
		Where("t.date <= ?", endDate).
		Where("t.deleted_at IS NULL")

	// Apply optional category filter
	if categoryID != nil {
		baseQuery = baseQuery.Where("t.category_id = ?", *categoryID)
	}

	// Count total
	var total int64
	countErr := baseQuery.Session(&gorm.Session{}).Count(&total).Error
	if countErr != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", countErr)
	}

	// Get paginated results
	err := baseQuery.
		Order("t.date DESC, t.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	transactions := make([]dashboard.PeriodTransaction, len(results))
	for i, res := range results {
		transactions[i] = dashboard.PeriodTransaction{
			ID:            res.ID,
			Description:   res.Description,
			Amount:        res.Amount,
			Date:          res.Date,
			CategoryID:    res.CategoryID,
			CategoryName:  res.CategoryName,
			CategoryColor: res.CategoryColor,
			CategoryIcon:  res.CategoryIcon,
		}
	}

	return transactions, int(total), nil
}

// GetPeriodSummary returns summary totals for a period.
func (r *dashboardRepository) GetPeriodSummary(
	ctx context.Context,
	userID uuid.UUID,
	startDate, endDate time.Time,
) (*dashboard.PeriodSummary, error) {
	var result struct {
		TotalIncome      decimal.Decimal `gorm:"column:total_income"`
		TotalExpenses    decimal.Decimal `gorm:"column:total_expenses"`
		Balance          decimal.Decimal `gorm:"column:balance"`
		TransactionCount int             `gorm:"column:transaction_count"`
	}

	query := `
		SELECT
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(amount), 0) as balance,
			COUNT(*) as transaction_count
		FROM transactions
		WHERE user_id = ?
			AND date >= ?
			AND date <= ?
			AND deleted_at IS NULL
	`

	err := r.db.WithContext(ctx).
		Raw(query, userID, startDate, endDate).
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get period summary: %w", err)
	}

	return &dashboard.PeriodSummary{
		TotalIncome:      result.TotalIncome,
		TotalExpenses:    result.TotalExpenses,
		Balance:          result.Balance,
		TransactionCount: result.TransactionCount,
	}, nil
}

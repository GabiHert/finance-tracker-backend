// Package dashboard contains dashboard-related use cases.
package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DashboardRepository defines the interface for dashboard data operations.
type DashboardRepository interface {
	// GetDateRange returns the date range of user's transactions.
	GetDateRange(ctx context.Context, userID uuid.UUID) (*DateRange, error)

	// GetAggregatedTrends returns income/expense trends aggregated by granularity.
	GetAggregatedTrends(
		ctx context.Context,
		userID uuid.UUID,
		startDate, endDate time.Time,
		granularity Granularity,
	) ([]RawTrendData, error)

	// GetCategoryBreakdown returns spending breakdown by category for a period.
	GetCategoryBreakdown(
		ctx context.Context,
		userID uuid.UUID,
		startDate, endDate time.Time,
	) ([]RawCategoryBreakdown, decimal.Decimal, error)

	// GetTransactionsByPeriod returns transactions for a specific period.
	GetTransactionsByPeriod(
		ctx context.Context,
		userID uuid.UUID,
		startDate, endDate time.Time,
		categoryID *uuid.UUID,
		limit, offset int,
	) ([]PeriodTransaction, int, error)

	// GetPeriodSummary returns summary totals for a period.
	GetPeriodSummary(
		ctx context.Context,
		userID uuid.UUID,
		startDate, endDate time.Time,
	) (*PeriodSummary, error)
}

// DateRange represents the date boundaries of a user's transaction history.
type DateRange struct {
	OldestDate        *time.Time
	NewestDate        *time.Time
	TotalTransactions int
}

// RawTrendData represents raw trend data from the database.
type RawTrendData struct {
	PeriodStart      time.Time
	Income           decimal.Decimal
	Expenses         decimal.Decimal
	TransactionCount int
}

// RawCategoryBreakdown represents raw category breakdown from the database.
type RawCategoryBreakdown struct {
	CategoryID       *uuid.UUID
	CategoryName     *string
	CategoryColor    *string
	CategoryIcon     *string
	Amount           decimal.Decimal
	TransactionCount int
}

// PeriodTransaction represents a transaction within a period.
type PeriodTransaction struct {
	ID            uuid.UUID
	Description   string
	Amount        decimal.Decimal
	Date          time.Time
	CategoryID    *uuid.UUID
	CategoryName  *string
	CategoryColor *string
	CategoryIcon  *string
}

// PeriodSummary represents summary totals for a period.
type PeriodSummary struct {
	TotalIncome      decimal.Decimal
	TotalExpenses    decimal.Decimal
	Balance          decimal.Decimal
	TransactionCount int
}

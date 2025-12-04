// Package group contains group-related use cases.
package group

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// DashboardPeriod represents the time period for dashboard data.
type DashboardPeriod string

const (
	PeriodThisMonth DashboardPeriod = "this_month"
	PeriodLastMonth DashboardPeriod = "last_month"
	PeriodThisWeek  DashboardPeriod = "this_week"
	PeriodLastWeek  DashboardPeriod = "last_week"
)

// GetGroupDashboardInput represents the input for getting group dashboard data.
type GetGroupDashboardInput struct {
	GroupID   uuid.UUID
	UserID    uuid.UUID
	Period    DashboardPeriod
	StartDate *time.Time // Optional: for custom date range
	EndDate   *time.Time // Optional: for custom date range
}

// GetGroupDashboardOutput represents the output of getting group dashboard data.
type GetGroupDashboardOutput struct {
	Dashboard *entity.GroupDashboardData
}

// GetGroupDashboardUseCase handles getting group dashboard data.
type GetGroupDashboardUseCase struct {
	groupRepo adapter.GroupRepository
}

// NewGetGroupDashboardUseCase creates a new GetGroupDashboardUseCase instance.
func NewGetGroupDashboardUseCase(groupRepo adapter.GroupRepository) *GetGroupDashboardUseCase {
	return &GetGroupDashboardUseCase{
		groupRepo: groupRepo,
	}
}

// Execute performs the group dashboard retrieval.
func (uc *GetGroupDashboardUseCase) Execute(ctx context.Context, input GetGroupDashboardInput) (*GetGroupDashboardOutput, error) {
	// Check if user is a member of the group
	isMember, err := uc.groupRepo.IsUserMemberOfGroup(ctx, input.GroupID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, domainerror.NewGroupError(
			domainerror.ErrCodeNotGroupMember,
			"you are not a member of this group",
			domainerror.ErrNotGroupMember,
		)
	}

	// Calculate date ranges - use custom dates if provided, otherwise use period
	var startDate, endDate time.Time
	var prevStartDate, prevEndDate time.Time

	if input.StartDate != nil && input.EndDate != nil {
		// Use custom date range
		startDate = *input.StartDate
		endDate = *input.EndDate
		// Calculate previous period as same duration, ending the day before start date
		prevStartDate, prevEndDate = uc.calculateCustomPreviousPeriod(startDate, endDate)
	} else {
		// Fall back to period-based calculation
		startDate, endDate = uc.calculateDateRange(input.Period)
		prevStartDate, prevEndDate = uc.calculatePreviousPeriodRange(input.Period, startDate, endDate)
	}

	// Get dashboard data for current period
	dashboard, err := uc.groupRepo.GetGroupDashboard(ctx, input.GroupID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard data: %w", err)
	}

	// Get previous period totals for comparison
	prevExpenses, prevIncome, err := uc.groupRepo.GetGroupDashboardPreviousPeriod(ctx, input.GroupID, prevStartDate, prevEndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous period data: %w", err)
	}

	// Calculate percent changes
	if dashboard.Summary != nil {
		dashboard.Summary.ExpensesChange = uc.calculatePercentChange(prevExpenses, dashboard.Summary.TotalExpenses)
		dashboard.Summary.IncomeChange = uc.calculatePercentChange(prevIncome, dashboard.Summary.TotalIncome)
	}

	return &GetGroupDashboardOutput{
		Dashboard: dashboard,
	}, nil
}

// calculateDateRange calculates the start and end dates based on the period.
func (uc *GetGroupDashboardUseCase) calculateDateRange(period DashboardPeriod) (time.Time, time.Time) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	switch period {
	case PeriodThisMonth:
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return startOfMonth, today

	case PeriodLastMonth:
		startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
		endOfLastMonth := startOfLastMonth.AddDate(0, 1, -1)
		return startOfLastMonth, endOfLastMonth

	case PeriodThisWeek:
		// Find most recent Monday
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday is 7
		}
		daysFromMonday := weekday - 1
		startOfWeek := today.AddDate(0, 0, -daysFromMonday)
		return startOfWeek, today

	case PeriodLastWeek:
		// Find the previous week's Monday and Sunday
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		daysFromMonday := weekday - 1
		thisWeekMonday := today.AddDate(0, 0, -daysFromMonday)
		lastWeekMonday := thisWeekMonday.AddDate(0, 0, -7)
		lastWeekSunday := thisWeekMonday.AddDate(0, 0, -1)
		return lastWeekMonday, lastWeekSunday

	default:
		// Default to this month
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return startOfMonth, today
	}
}

// calculatePreviousPeriodRange calculates the previous period's date range for comparison.
func (uc *GetGroupDashboardUseCase) calculatePreviousPeriodRange(period DashboardPeriod, currentStart, currentEnd time.Time) (time.Time, time.Time) {
	switch period {
	case PeriodThisMonth:
		// Compare to same days in the previous month
		prevStart := currentStart.AddDate(0, -1, 0)
		duration := currentEnd.Sub(currentStart)
		prevEnd := prevStart.Add(duration)
		return prevStart, prevEnd

	case PeriodLastMonth:
		// Compare to the month before last month
		prevStart := currentStart.AddDate(0, -1, 0)
		prevEnd := prevStart.AddDate(0, 1, -1)
		return prevStart, prevEnd

	case PeriodThisWeek:
		// Compare to last week
		prevStart := currentStart.AddDate(0, 0, -7)
		duration := currentEnd.Sub(currentStart)
		prevEnd := prevStart.Add(duration)
		return prevStart, prevEnd

	case PeriodLastWeek:
		// Compare to the week before last week
		prevStart := currentStart.AddDate(0, 0, -7)
		prevEnd := currentEnd.AddDate(0, 0, -7)
		return prevStart, prevEnd

	default:
		// Default: previous month
		prevStart := currentStart.AddDate(0, -1, 0)
		duration := currentEnd.Sub(currentStart)
		prevEnd := prevStart.Add(duration)
		return prevStart, prevEnd
	}
}

// calculatePercentChange calculates the percent change between two values.
func (uc *GetGroupDashboardUseCase) calculatePercentChange(previous, current float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100 // Infinite increase represented as 100%
	}
	return ((current - previous) / previous) * 100
}

// calculateCustomPreviousPeriod calculates the previous period for a custom date range.
// The previous period has the same duration and ends the day before the current period starts.
func (uc *GetGroupDashboardUseCase) calculateCustomPreviousPeriod(startDate, endDate time.Time) (time.Time, time.Time) {
	// Calculate duration of the custom period (inclusive of both start and end dates)
	duration := endDate.Sub(startDate)

	// Previous period ends the day before the current period starts
	prevEndDate := startDate.AddDate(0, 0, -1)
	// Previous period has the same duration
	prevStartDate := prevEndDate.Add(-duration)

	return prevStartDate, prevEndDate
}

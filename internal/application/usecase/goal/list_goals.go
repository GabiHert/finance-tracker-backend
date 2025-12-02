// Package goal contains goal-related use cases.
package goal

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// ListGoalsInput represents the input for listing goals.
type ListGoalsInput struct {
	UserID uuid.UUID
}

// ListGoalsOutput represents the output of listing goals.
type ListGoalsOutput struct {
	Goals []*GoalOutput
}

// GoalOutput represents a single goal in the output.
type GoalOutput struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	CategoryID    uuid.UUID
	Category      *entity.Category
	LimitAmount   float64
	CurrentAmount float64
	AlertOnExceed bool
	Period        entity.GoalPeriod
	StartDate     *time.Time
	EndDate       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ListGoalsUseCase handles listing goals logic.
type ListGoalsUseCase struct {
	goalRepo     adapter.GoalRepository
	categoryRepo adapter.CategoryRepository
}

// NewListGoalsUseCase creates a new ListGoalsUseCase instance.
func NewListGoalsUseCase(goalRepo adapter.GoalRepository, categoryRepo adapter.CategoryRepository) *ListGoalsUseCase {
	return &ListGoalsUseCase{
		goalRepo:     goalRepo,
		categoryRepo: categoryRepo,
	}
}

// Execute performs the goal listing.
func (uc *ListGoalsUseCase) Execute(ctx context.Context, input ListGoalsInput) (*ListGoalsOutput, error) {
	goals, err := uc.goalRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	output := &ListGoalsOutput{
		Goals: make([]*GoalOutput, 0, len(goals)),
	}

	for _, g := range goals {
		// Fetch category for this goal
		cat, err := uc.categoryRepo.FindByID(ctx, g.CategoryID)
		if err != nil {
			// If category not found, skip this goal or continue without category
			cat = nil
		}

		// Calculate period dates
		startDate, endDate := calculatePeriodDates(g.Period, g.StartDate, g.EndDate)

		// Get current spending for this category within the period
		currentAmount, err := uc.goalRepo.GetCurrentSpending(ctx, g.CategoryID, startDate, endDate)
		if err != nil {
			currentAmount = 0
		}

		goalOutput := &GoalOutput{
			ID:            g.ID,
			UserID:        g.UserID,
			CategoryID:    g.CategoryID,
			Category:      cat,
			LimitAmount:   g.LimitAmount,
			CurrentAmount: currentAmount,
			AlertOnExceed: g.AlertOnExceed,
			Period:        g.Period,
			StartDate:     g.StartDate,
			EndDate:       g.EndDate,
			CreatedAt:     g.CreatedAt,
			UpdatedAt:     g.UpdatedAt,
		}

		output.Goals = append(output.Goals, goalOutput)
	}

	return output, nil
}

// calculatePeriodDates calculates the start and end dates for a goal period.
func calculatePeriodDates(period entity.GoalPeriod, customStart, customEnd *time.Time) (time.Time, time.Time) {
	now := time.Now().UTC()

	// If custom dates are provided, use them
	if customStart != nil && customEnd != nil {
		return *customStart, *customEnd
	}

	switch period {
	case entity.GoalPeriodWeekly:
		// Start from the beginning of the current week (Sunday)
		weekday := int(now.Weekday())
		startDate := now.AddDate(0, 0, -weekday)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 0, 6)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)
		return startDate, endDate

	case entity.GoalPeriodYearly:
		// Start from the beginning of the current year
		startDate := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(now.Year(), 12, 31, 23, 59, 59, 999999999, time.UTC)
		return startDate, endDate

	case entity.GoalPeriodMonthly:
		fallthrough
	default:
		// Start from the beginning of the current month
		startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, -1)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)
		return startDate, endDate
	}
}

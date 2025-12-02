// Package goal contains goal-related use cases.
package goal

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// GetGoalInput represents the input for getting a goal.
type GetGoalInput struct {
	GoalID uuid.UUID
	UserID uuid.UUID
}

// GetGoalOutput represents the output of getting a goal.
type GetGoalOutput struct {
	Goal          *entity.Goal
	Category      *entity.Category
	CurrentAmount float64
}

// GetGoalUseCase handles getting a goal by ID.
type GetGoalUseCase struct {
	goalRepo     adapter.GoalRepository
	categoryRepo adapter.CategoryRepository
}

// NewGetGoalUseCase creates a new GetGoalUseCase instance.
func NewGetGoalUseCase(goalRepo adapter.GoalRepository, categoryRepo adapter.CategoryRepository) *GetGoalUseCase {
	return &GetGoalUseCase{
		goalRepo:     goalRepo,
		categoryRepo: categoryRepo,
	}
}

// Execute performs the goal retrieval.
func (uc *GetGoalUseCase) Execute(ctx context.Context, input GetGoalInput) (*GetGoalOutput, error) {
	// Find the goal
	goal, err := uc.goalRepo.FindByID(ctx, input.GoalID)
	if err != nil {
		if errors.Is(err, domainerror.ErrGoalNotFound) {
			return nil, domainerror.NewGoalError(
				domainerror.ErrCodeGoalNotFound,
				"goal not found",
				domainerror.ErrGoalNotFound,
			)
		}
		return nil, fmt.Errorf("failed to find goal: %w", err)
	}

	// Check if user is authorized to access this goal
	if goal.UserID != input.UserID {
		return nil, domainerror.NewGoalError(
			domainerror.ErrCodeUnauthorizedGoalAccess,
			"not authorized to access this goal",
			domainerror.ErrUnauthorizedGoalAccess,
		)
	}

	// Fetch category for this goal
	category, err := uc.categoryRepo.FindByID(ctx, goal.CategoryID)
	if err != nil {
		category = nil
	}

	// Calculate period dates
	startDate, endDate := calculatePeriodDates(goal.Period, goal.StartDate, goal.EndDate)

	// Get current spending for this category within the period
	currentAmount, err := uc.goalRepo.GetCurrentSpending(ctx, goal.CategoryID, startDate, endDate)
	if err != nil {
		currentAmount = 0
	}

	return &GetGoalOutput{
		Goal:          goal,
		Category:      category,
		CurrentAmount: currentAmount,
	}, nil
}

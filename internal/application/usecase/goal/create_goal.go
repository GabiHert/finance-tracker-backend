// Package goal contains goal-related use cases.
package goal

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// CreateGoalInput represents the input for goal creation.
type CreateGoalInput struct {
	UserID        uuid.UUID
	CategoryID    uuid.UUID
	LimitAmount   float64
	AlertOnExceed *bool              // Optional, defaults to true
	Period        *entity.GoalPeriod // Optional, defaults to monthly
}

// CreateGoalOutput represents the output of goal creation.
type CreateGoalOutput struct {
	Goal *entity.Goal
}

// CreateGoalUseCase handles goal creation logic.
type CreateGoalUseCase struct {
	goalRepo     adapter.GoalRepository
	categoryRepo adapter.CategoryRepository
}

// NewCreateGoalUseCase creates a new CreateGoalUseCase instance.
func NewCreateGoalUseCase(goalRepo adapter.GoalRepository, categoryRepo adapter.CategoryRepository) *CreateGoalUseCase {
	return &CreateGoalUseCase{
		goalRepo:     goalRepo,
		categoryRepo: categoryRepo,
	}
}

// Execute performs the goal creation.
func (uc *CreateGoalUseCase) Execute(ctx context.Context, input CreateGoalInput) (*CreateGoalOutput, error) {
	// Validate limit amount
	if input.LimitAmount <= 0 {
		return nil, domainerror.NewGoalError(
			domainerror.ErrCodeInvalidLimitAmount,
			"limit amount must be greater than zero",
			domainerror.ErrInvalidLimitAmount,
		)
	}

	// Validate category exists
	category, err := uc.categoryRepo.FindByID(ctx, input.CategoryID)
	if err != nil {
		return nil, domainerror.NewGoalError(
			domainerror.ErrCodeGoalCategoryNotFound,
			"category not found",
			domainerror.ErrGoalCategoryNotFound,
		)
	}

	// Validate category belongs to user
	if category.OwnerType != entity.OwnerTypeUser || category.OwnerID != input.UserID {
		return nil, domainerror.NewGoalError(
			domainerror.ErrCodeCategoryDoesNotBelongUser,
			"category does not belong to user",
			domainerror.ErrCategoryDoesNotBelongToUser,
		)
	}

	// Check if goal already exists for this category
	exists, err := uc.goalRepo.ExistsByUserAndCategory(ctx, input.UserID, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check goal existence: %w", err)
	}
	if exists {
		return nil, domainerror.NewGoalError(
			domainerror.ErrCodeGoalAlreadyExists,
			"a goal already exists for this category",
			domainerror.ErrGoalAlreadyExists,
		)
	}

	// Apply defaults
	alertOnExceed := true
	if input.AlertOnExceed != nil {
		alertOnExceed = *input.AlertOnExceed
	}

	period := entity.GoalPeriodMonthly
	if input.Period != nil {
		// Validate period
		if !isValidGoalPeriod(*input.Period) {
			return nil, domainerror.NewGoalError(
				domainerror.ErrCodeInvalidGoalPeriod,
				"period must be 'monthly', 'weekly', or 'yearly'",
				domainerror.ErrInvalidGoalPeriod,
			)
		}
		period = *input.Period
	}

	// Create goal entity
	goal := entity.NewGoal(
		input.UserID,
		input.CategoryID,
		input.LimitAmount,
		alertOnExceed,
		period,
	)

	// Save goal to database
	if err := uc.goalRepo.Create(ctx, goal); err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	return &CreateGoalOutput{
		Goal: goal,
	}, nil
}

// isValidGoalPeriod validates the goal period.
func isValidGoalPeriod(period entity.GoalPeriod) bool {
	return period == entity.GoalPeriodMonthly ||
		period == entity.GoalPeriodWeekly ||
		period == entity.GoalPeriodYearly
}

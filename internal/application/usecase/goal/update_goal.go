// Package goal contains goal-related use cases.
package goal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// UpdateGoalInput represents the input for goal update.
type UpdateGoalInput struct {
	GoalID        uuid.UUID
	UserID        uuid.UUID
	LimitAmount   *float64           // Optional
	AlertOnExceed *bool              // Optional
	Period        *entity.GoalPeriod // Optional
}

// UpdateGoalOutput represents the output of goal update.
type UpdateGoalOutput struct {
	Goal *entity.Goal
}

// UpdateGoalUseCase handles goal update logic.
type UpdateGoalUseCase struct {
	goalRepo adapter.GoalRepository
}

// NewUpdateGoalUseCase creates a new UpdateGoalUseCase instance.
func NewUpdateGoalUseCase(goalRepo adapter.GoalRepository) *UpdateGoalUseCase {
	return &UpdateGoalUseCase{
		goalRepo: goalRepo,
	}
}

// Execute performs the goal update.
func (uc *UpdateGoalUseCase) Execute(ctx context.Context, input UpdateGoalInput) (*UpdateGoalOutput, error) {
	// Find the existing goal
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

	// Check if user is authorized to modify this goal
	if goal.UserID != input.UserID {
		return nil, domainerror.NewGoalError(
			domainerror.ErrCodeUnauthorizedGoalAccess,
			"not authorized to modify this goal",
			domainerror.ErrUnauthorizedGoalAccess,
		)
	}

	// Update limit amount if provided
	if input.LimitAmount != nil {
		if *input.LimitAmount <= 0 {
			return nil, domainerror.NewGoalError(
				domainerror.ErrCodeInvalidLimitAmount,
				"limit amount must be greater than zero",
				domainerror.ErrInvalidLimitAmount,
			)
		}
		goal.LimitAmount = *input.LimitAmount
	}

	// Update alert_on_exceed if provided
	if input.AlertOnExceed != nil {
		goal.AlertOnExceed = *input.AlertOnExceed
	}

	// Update period if provided
	if input.Period != nil {
		if !isValidGoalPeriod(*input.Period) {
			return nil, domainerror.NewGoalError(
				domainerror.ErrCodeInvalidGoalPeriod,
				"period must be 'monthly', 'weekly', or 'yearly'",
				domainerror.ErrInvalidGoalPeriod,
			)
		}
		goal.Period = *input.Period
	}

	// Update timestamp
	goal.UpdatedAt = time.Now().UTC()

	// Save updated goal
	if err := uc.goalRepo.Update(ctx, goal); err != nil {
		return nil, fmt.Errorf("failed to update goal: %w", err)
	}

	return &UpdateGoalOutput{
		Goal: goal,
	}, nil
}

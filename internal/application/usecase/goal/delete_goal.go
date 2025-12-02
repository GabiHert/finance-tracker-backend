// Package goal contains goal-related use cases.
package goal

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// DeleteGoalInput represents the input for goal deletion.
type DeleteGoalInput struct {
	GoalID uuid.UUID
	UserID uuid.UUID
}

// DeleteGoalOutput represents the output of goal deletion.
type DeleteGoalOutput struct {
	Success bool
}

// DeleteGoalUseCase handles goal deletion logic.
type DeleteGoalUseCase struct {
	goalRepo adapter.GoalRepository
}

// NewDeleteGoalUseCase creates a new DeleteGoalUseCase instance.
func NewDeleteGoalUseCase(goalRepo adapter.GoalRepository) *DeleteGoalUseCase {
	return &DeleteGoalUseCase{
		goalRepo: goalRepo,
	}
}

// Execute performs the goal deletion.
func (uc *DeleteGoalUseCase) Execute(ctx context.Context, input DeleteGoalInput) (*DeleteGoalOutput, error) {
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

	// Check if user is authorized to delete this goal
	if goal.UserID != input.UserID {
		return nil, domainerror.NewGoalError(
			domainerror.ErrCodeUnauthorizedGoalAccess,
			"not authorized to delete this goal",
			domainerror.ErrUnauthorizedGoalAccess,
		)
	}

	// Delete the goal
	if err := uc.goalRepo.Delete(ctx, input.GoalID); err != nil {
		return nil, fmt.Errorf("failed to delete goal: %w", err)
	}

	return &DeleteGoalOutput{
		Success: true,
	}, nil
}

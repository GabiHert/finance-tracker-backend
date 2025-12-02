// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// GoalRepository defines the interface for goal persistence operations.
type GoalRepository interface {
	// Create creates a new goal in the database.
	Create(ctx context.Context, goal *entity.Goal) error

	// FindByID retrieves a goal by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Goal, error)

	// FindByUserID retrieves all goals for a given user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Goal, error)

	// FindByUserAndCategory retrieves a goal by user ID and category ID.
	FindByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) (*entity.Goal, error)

	// Update updates an existing goal in the database.
	Update(ctx context.Context, goal *entity.Goal) error

	// Delete removes a goal from the database (soft delete).
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByUserAndCategory checks if a goal exists for the given user and category.
	ExistsByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) (bool, error)

	// GetCurrentSpending calculates the current spending for a category within the goal period.
	GetCurrentSpending(ctx context.Context, categoryID uuid.UUID, startDate, endDate time.Time) (float64, error)
}

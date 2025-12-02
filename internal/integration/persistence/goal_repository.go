// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

// goalRepository implements the adapter.GoalRepository interface.
type goalRepository struct {
	db *gorm.DB
}

// NewGoalRepository creates a new goal repository instance.
func NewGoalRepository(db *gorm.DB) adapter.GoalRepository {
	return &goalRepository{
		db: db,
	}
}

// Create creates a new goal in the database.
func (r *goalRepository) Create(ctx context.Context, goal *entity.Goal) error {
	goalModel := model.GoalFromEntity(goal)
	result := r.db.WithContext(ctx).Create(goalModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID retrieves a goal by its ID.
func (r *goalRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Goal, error) {
	var goalModel model.GoalModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&goalModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrGoalNotFound
		}
		return nil, result.Error
	}
	return goalModel.ToEntity(), nil
}

// FindByUserID retrieves all goals for a given user.
func (r *goalRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Goal, error) {
	var goalModels []model.GoalModel
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&goalModels)
	if result.Error != nil {
		return nil, result.Error
	}

	goals := make([]*entity.Goal, len(goalModels))
	for i, gm := range goalModels {
		goals[i] = gm.ToEntity()
	}
	return goals, nil
}

// FindByUserAndCategory retrieves a goal by user ID and category ID.
func (r *goalRepository) FindByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) (*entity.Goal, error) {
	var goalModel model.GoalModel
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND category_id = ?", userID, categoryID).
		First(&goalModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return goalModel.ToEntity(), nil
}

// Update updates an existing goal in the database.
func (r *goalRepository) Update(ctx context.Context, goal *entity.Goal) error {
	goalModel := model.GoalFromEntity(goal)
	result := r.db.WithContext(ctx).Save(goalModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete removes a goal from the database (soft delete).
func (r *goalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.GoalModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ExistsByUserAndCategory checks if a goal exists for the given user and category.
func (r *goalRepository) ExistsByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.GoalModel{}).
		Where("user_id = ? AND category_id = ?", userID, categoryID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// GetCurrentSpending calculates the current spending for a category within the goal period.
func (r *goalRepository) GetCurrentSpending(ctx context.Context, categoryID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var total float64

	// Query transactions for this category within the date range
	// Only count expense transactions (negative amounts or expense type categories)
	result := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Select("COALESCE(SUM(ABS(amount)), 0)").
		Where("category_id = ?", categoryID).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Scan(&total)

	if result.Error != nil {
		return 0, result.Error
	}

	return total, nil
}

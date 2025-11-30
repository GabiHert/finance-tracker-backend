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

// categoryRepository implements the adapter.CategoryRepository interface.
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository instance.
func NewCategoryRepository(db *gorm.DB) adapter.CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

// Create creates a new category in the database.
func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	categoryModel := model.CategoryFromEntity(category)
	result := r.db.WithContext(ctx).Create(categoryModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID retrieves a category by its ID.
func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var categoryModel model.CategoryModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&categoryModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrCategoryNotFound
		}
		return nil, result.Error
	}
	return categoryModel.ToEntity(), nil
}

// FindByOwner retrieves all categories for a given owner.
func (r *categoryRepository) FindByOwner(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) ([]*entity.Category, error) {
	var categoryModels []model.CategoryModel
	result := r.db.WithContext(ctx).
		Where("owner_type = ? AND owner_id = ?", string(ownerType), ownerID).
		Order("name ASC").
		Find(&categoryModels)
	if result.Error != nil {
		return nil, result.Error
	}

	categories := make([]*entity.Category, len(categoryModels))
	for i, cm := range categoryModels {
		categories[i] = cm.ToEntity()
	}
	return categories, nil
}

// FindByOwnerAndType retrieves categories for a given owner filtered by type.
func (r *categoryRepository) FindByOwnerAndType(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID, categoryType entity.CategoryType) ([]*entity.Category, error) {
	var categoryModels []model.CategoryModel
	result := r.db.WithContext(ctx).
		Where("owner_type = ? AND owner_id = ? AND type = ?", string(ownerType), ownerID, string(categoryType)).
		Order("name ASC").
		Find(&categoryModels)
	if result.Error != nil {
		return nil, result.Error
	}

	categories := make([]*entity.Category, len(categoryModels))
	for i, cm := range categoryModels {
		categories[i] = cm.ToEntity()
	}
	return categories, nil
}

// FindByNameAndOwner retrieves a category by name and owner.
func (r *categoryRepository) FindByNameAndOwner(ctx context.Context, name string, ownerType entity.OwnerType, ownerID uuid.UUID) (*entity.Category, error) {
	var categoryModel model.CategoryModel
	result := r.db.WithContext(ctx).
		Where("name = ? AND owner_type = ? AND owner_id = ?", name, string(ownerType), ownerID).
		First(&categoryModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return categoryModel.ToEntity(), nil
}

// Update updates an existing category in the database.
func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	categoryModel := model.CategoryFromEntity(category)
	result := r.db.WithContext(ctx).Save(categoryModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete removes a category from the database.
func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.CategoryModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ExistsByNameAndOwner checks if a category with the given name exists for the owner.
func (r *categoryRepository) ExistsByNameAndOwner(ctx context.Context, name string, ownerType entity.OwnerType, ownerID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.CategoryModel{}).
		Where("name = ? AND owner_type = ? AND owner_id = ?", name, string(ownerType), ownerID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// categoryStatsResult represents the result of a category statistics query.
type categoryStatsResult struct {
	CategoryID       uuid.UUID
	TransactionCount int
	PeriodTotal      float64
}

// GetTransactionStats retrieves transaction statistics for categories within a date range.
func (r *categoryRepository) GetTransactionStats(ctx context.Context, categoryIDs []uuid.UUID, startDate, endDate time.Time) (map[uuid.UUID]*adapter.CategoryStats, error) {
	// Initialize stats map with zeros for all requested category IDs
	stats := make(map[uuid.UUID]*adapter.CategoryStats)
	for _, id := range categoryIDs {
		stats[id] = &adapter.CategoryStats{
			TransactionCount: 0,
			PeriodTotal:      0,
		}
	}

	// Return early if no category IDs provided
	if len(categoryIDs) == 0 {
		return stats, nil
	}

	// Query transaction statistics grouped by category_id
	var results []categoryStatsResult
	query := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Select("category_id, COUNT(*) as transaction_count, COALESCE(SUM(amount), 0) as period_total").
		Where("category_id IN ?", categoryIDs).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Group("category_id")

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	// Update stats map with query results
	for _, result := range results {
		if _, exists := stats[result.CategoryID]; exists {
			stats[result.CategoryID].TransactionCount = result.TransactionCount
			stats[result.CategoryID].PeriodTotal = result.PeriodTotal
		}
	}

	return stats, nil
}

// OrphanTransactionsByCategory sets category_id to NULL for all transactions with the given category ID.
func (r *categoryRepository) OrphanTransactionsByCategory(ctx context.Context, categoryID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("category_id = ?", categoryID).
		Update("category_id", nil)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

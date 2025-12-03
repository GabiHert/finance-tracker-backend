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

// categoryRuleRepository implements the adapter.CategoryRuleRepository interface.
type categoryRuleRepository struct {
	db *gorm.DB
}

// NewCategoryRuleRepository creates a new category rule repository instance.
func NewCategoryRuleRepository(db *gorm.DB) adapter.CategoryRuleRepository {
	return &categoryRuleRepository{
		db: db,
	}
}

// Create creates a new category rule in the database.
func (r *categoryRuleRepository) Create(ctx context.Context, rule *entity.CategoryRule) error {
	ruleModel := model.CategoryRuleFromEntity(rule)
	result := r.db.WithContext(ctx).Create(ruleModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID retrieves a category rule by its ID.
func (r *categoryRuleRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.CategoryRule, error) {
	var ruleModel model.CategoryRuleModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&ruleModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrCategoryRuleNotFound
		}
		return nil, result.Error
	}
	return ruleModel.ToEntity(), nil
}

// FindByIDWithCategory retrieves a category rule with its category by ID.
func (r *categoryRuleRepository) FindByIDWithCategory(ctx context.Context, id uuid.UUID) (*entity.CategoryRuleWithCategory, error) {
	var ruleModel model.CategoryRuleModel
	result := r.db.WithContext(ctx).
		Preload("Category").
		Where("id = ?", id).
		First(&ruleModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrCategoryRuleNotFound
		}
		return nil, result.Error
	}
	return ruleModel.ToEntityWithCategory(), nil
}

// FindByOwner retrieves all category rules for a given owner, sorted by priority (descending).
func (r *categoryRuleRepository) FindByOwner(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) ([]*entity.CategoryRule, error) {
	var ruleModels []model.CategoryRuleModel
	result := r.db.WithContext(ctx).
		Where("owner_type = ? AND owner_id = ?", string(ownerType), ownerID).
		Order("priority DESC").
		Find(&ruleModels)
	if result.Error != nil {
		return nil, result.Error
	}

	rules := make([]*entity.CategoryRule, len(ruleModels))
	for i, rm := range ruleModels {
		rules[i] = rm.ToEntity()
	}
	return rules, nil
}

// FindByOwnerWithCategories retrieves all category rules with their categories for a given owner.
func (r *categoryRuleRepository) FindByOwnerWithCategories(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) ([]*entity.CategoryRuleWithCategory, error) {
	var ruleModels []model.CategoryRuleModel
	result := r.db.WithContext(ctx).
		Preload("Category").
		Where("owner_type = ? AND owner_id = ?", string(ownerType), ownerID).
		Order("priority DESC").
		Find(&ruleModels)
	if result.Error != nil {
		return nil, result.Error
	}

	rules := make([]*entity.CategoryRuleWithCategory, len(ruleModels))
	for i, rm := range ruleModels {
		rules[i] = rm.ToEntityWithCategory()
	}
	return rules, nil
}

// FindActiveByOwner retrieves only active category rules for a given owner, sorted by priority (descending).
func (r *categoryRuleRepository) FindActiveByOwner(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) ([]*entity.CategoryRule, error) {
	var ruleModels []model.CategoryRuleModel
	result := r.db.WithContext(ctx).
		Where("owner_type = ? AND owner_id = ? AND is_active = ?", string(ownerType), ownerID, true).
		Order("priority DESC").
		Find(&ruleModels)
	if result.Error != nil {
		return nil, result.Error
	}

	rules := make([]*entity.CategoryRule, len(ruleModels))
	for i, rm := range ruleModels {
		rules[i] = rm.ToEntity()
	}
	return rules, nil
}

// Update updates an existing category rule in the database.
func (r *categoryRuleRepository) Update(ctx context.Context, rule *entity.CategoryRule) error {
	ruleModel := model.CategoryRuleFromEntity(rule)
	result := r.db.WithContext(ctx).Save(ruleModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete removes a category rule from the database (hard delete).
// Using Unscoped() to bypass soft-delete and permanently remove the record.
// This allows the same pattern to be reused after deletion.
func (r *categoryRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&model.CategoryRuleModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ExistsByPatternAndOwner checks if a rule with the given pattern exists for the owner.
func (r *categoryRuleRepository) ExistsByPatternAndOwner(ctx context.Context, pattern string, ownerType entity.OwnerType, ownerID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.CategoryRuleModel{}).
		Where("pattern = ? AND owner_type = ? AND owner_id = ?", pattern, string(ownerType), ownerID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// ExistsByPatternAndOwnerExcluding checks if a rule with the given pattern exists for the owner,
// excluding a specific rule ID (used for updates).
func (r *categoryRuleRepository) ExistsByPatternAndOwnerExcluding(ctx context.Context, pattern string, ownerType entity.OwnerType, ownerID uuid.UUID, excludeID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.CategoryRuleModel{}).
		Where("pattern = ? AND owner_type = ? AND owner_id = ? AND id != ?", pattern, string(ownerType), ownerID, excludeID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// UpdatePriorities updates the priorities for multiple rules in a batch operation.
func (r *categoryRuleRepository) UpdatePriorities(ctx context.Context, updates []entity.RulePriorityUpdate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		for _, update := range updates {
			result := tx.Model(&model.CategoryRuleModel{}).
				Where("id = ?", update.ID).
				Updates(map[string]interface{}{
					"priority":   update.Priority,
					"updated_at": now,
				})
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
}

// matchingTransactionResult represents a raw query result for matching transactions.
type matchingTransactionResult struct {
	ID          uuid.UUID
	Description string
	Amount      string
	Date        time.Time
}

// FindMatchingTransactions finds transactions that match the given regex pattern.
// Uses PostgreSQL's ~ operator for regex matching.
func (r *categoryRuleRepository) FindMatchingTransactions(ctx context.Context, pattern string, ownerType entity.OwnerType, ownerID uuid.UUID, limit int) (*entity.PatternTestResult, error) {
	// Build the query based on owner type
	// For user ownership, we match on user_id
	// For group ownership, we need to find transactions of users who are members of the group
	// For simplicity, we'll start with user ownership

	var results []matchingTransactionResult
	var totalCount int64

	// Query for user-owned rules - match against transactions owned by the user
	if ownerType == entity.OwnerTypeUser {
		// Count total matches
		countQuery := r.db.WithContext(ctx).
			Table("transactions").
			Where("user_id = ? AND description ~* ? AND deleted_at IS NULL", ownerID, pattern)

		if err := countQuery.Count(&totalCount).Error; err != nil {
			return nil, err
		}

		// Get matching transactions with limit
		query := r.db.WithContext(ctx).
			Table("transactions").
			Select("id, description, amount::text as amount, date").
			Where("user_id = ? AND description ~* ? AND deleted_at IS NULL", ownerID, pattern).
			Order("date DESC").
			Limit(limit)

		if err := query.Scan(&results).Error; err != nil {
			return nil, err
		}
	} else {
		// For group ownership, we need to find transactions of group members
		// This is a simplified implementation - may need refinement
		countQuery := r.db.WithContext(ctx).
			Table("transactions t").
			Joins("INNER JOIN group_members gm ON t.user_id = gm.user_id").
			Where("gm.group_id = ? AND t.description ~* ? AND t.deleted_at IS NULL", ownerID, pattern)

		if err := countQuery.Count(&totalCount).Error; err != nil {
			return nil, err
		}

		query := r.db.WithContext(ctx).
			Table("transactions t").
			Select("t.id, t.description, t.amount::text as amount, t.date").
			Joins("INNER JOIN group_members gm ON t.user_id = gm.user_id").
			Where("gm.group_id = ? AND t.description ~* ? AND t.deleted_at IS NULL", ownerID, pattern).
			Order("t.date DESC").
			Limit(limit)

		if err := query.Scan(&results).Error; err != nil {
			return nil, err
		}
	}

	// Convert results to domain entities
	matchingTxs := make([]*entity.MatchingTransaction, len(results))
	for i, result := range results {
		matchingTxs[i] = &entity.MatchingTransaction{
			ID:          result.ID,
			Description: result.Description,
			Amount:      result.Amount,
			Date:        result.Date,
		}
	}

	return &entity.PatternTestResult{
		MatchingTransactions: matchingTxs,
		MatchCount:           int(totalCount),
	}, nil
}

// GetMaxPriorityByOwner gets the maximum priority value for rules owned by the given owner.
func (r *categoryRuleRepository) GetMaxPriorityByOwner(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) (int, error) {
	var maxPriority *int
	result := r.db.WithContext(ctx).
		Model(&model.CategoryRuleModel{}).
		Select("COALESCE(MAX(priority), 0)").
		Where("owner_type = ? AND owner_id = ?", string(ownerType), ownerID).
		Scan(&maxPriority)

	if result.Error != nil {
		return 0, result.Error
	}

	if maxPriority == nil {
		return 0, nil
	}
	return *maxPriority, nil
}

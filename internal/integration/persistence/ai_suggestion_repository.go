// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

// aiSuggestionRepository implements the adapter.AISuggestionRepository interface.
type aiSuggestionRepository struct {
	db *gorm.DB
}

// NewAISuggestionRepository creates a new AI suggestion repository instance.
func NewAISuggestionRepository(db *gorm.DB) adapter.AISuggestionRepository {
	return &aiSuggestionRepository{
		db: db,
	}
}

// Create creates a new AI suggestion in the database.
func (r *aiSuggestionRepository) Create(ctx context.Context, suggestion *entity.AISuggestion) error {
	suggestionModel := model.AISuggestionFromEntity(suggestion)
	result := r.db.WithContext(ctx).Create(suggestionModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// CreateBatch creates multiple AI suggestions in a single transaction.
func (r *aiSuggestionRepository) CreateBatch(ctx context.Context, suggestions []*entity.AISuggestion) error {
	if len(suggestions) == 0 {
		return nil
	}

	models := make([]*model.AISuggestionModel, len(suggestions))
	for i, s := range suggestions {
		models[i] = model.AISuggestionFromEntity(s)
	}

	result := r.db.WithContext(ctx).CreateInBatches(models, 100)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID retrieves an AI suggestion by its ID.
func (r *aiSuggestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AISuggestion, error) {
	var suggestionModel model.AISuggestionModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&suggestionModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrAISuggestionNotFound
		}
		return nil, result.Error
	}
	return suggestionModel.ToEntity(), nil
}

// GetByIDWithDetails retrieves an AI suggestion with all related details.
func (r *aiSuggestionRepository) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.AISuggestionWithDetails, error) {
	var suggestionModel model.AISuggestionModel
	result := r.db.WithContext(ctx).
		Preload("Transaction").
		Preload("Category").
		Where("id = ?", id).
		First(&suggestionModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrAISuggestionNotFound
		}
		return nil, result.Error
	}
	return suggestionModel.ToEntityWithDetails(), nil
}

// GetByUserID retrieves all AI suggestions for a given user.
func (r *aiSuggestionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.AISuggestion, error) {
	var suggestionModels []model.AISuggestionModel
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&suggestionModels)
	if result.Error != nil {
		return nil, result.Error
	}

	suggestions := make([]*entity.AISuggestion, len(suggestionModels))
	for i, sm := range suggestionModels {
		suggestions[i] = sm.ToEntity()
	}
	return suggestions, nil
}

// GetPendingByUserID retrieves all pending AI suggestions for a given user with details.
func (r *aiSuggestionRepository) GetPendingByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.AISuggestionWithDetails, error) {
	var suggestionModels []model.AISuggestionModel
	result := r.db.WithContext(ctx).
		Preload("Transaction").
		Preload("Category").
		Where("user_id = ? AND status = ?", userID, string(entity.SuggestionStatusPending)).
		Order("created_at DESC").
		Find(&suggestionModels)
	if result.Error != nil {
		return nil, result.Error
	}

	// Collect all unique affected transaction IDs from all suggestions
	txnIDSet := make(map[uuid.UUID]struct{})
	for _, sm := range suggestionModels {
		for _, idStr := range sm.AffectedTransactionIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				txnIDSet[id] = struct{}{}
			}
		}
	}

	// Fetch all affected transactions in one batch query
	txnMap := make(map[uuid.UUID]*model.TransactionModel)
	if len(txnIDSet) > 0 {
		allTxnIDs := make([]uuid.UUID, 0, len(txnIDSet))
		for id := range txnIDSet {
			allTxnIDs = append(allTxnIDs, id)
		}

		var transactions []model.TransactionModel
		if err := r.db.WithContext(ctx).
			Where("id IN ?", allTxnIDs).
			Find(&transactions).Error; err != nil {
			// Log error but continue - some transactions might be deleted
		}

		for i := range transactions {
			txnMap[transactions[i].ID] = &transactions[i]
		}
	}

	// Build results with populated affected transactions
	suggestions := make([]*entity.AISuggestionWithDetails, len(suggestionModels))
	for i := range suggestionModels {
		suggestions[i] = suggestionModels[i].ToEntityWithDetailsAndTransactions(txnMap)
	}
	return suggestions, nil
}

// GetPendingCount retrieves the count of pending AI suggestions for a given user.
func (r *aiSuggestionRepository) GetPendingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.AISuggestionModel{}).
		Where("user_id = ? AND status = ?", userID, string(entity.SuggestionStatusPending)).
		Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

// Update updates an existing AI suggestion in the database.
func (r *aiSuggestionRepository) Update(ctx context.Context, suggestion *entity.AISuggestion) error {
	suggestionModel := model.AISuggestionFromEntity(suggestion)
	result := r.db.WithContext(ctx).Save(suggestionModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteByUserID deletes all AI suggestions for a given user.
func (r *aiSuggestionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	result := r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ?", userID).
		Delete(&model.AISuggestionModel{})
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

// DeleteByID deletes an AI suggestion by its ID.
func (r *aiSuggestionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&model.AISuggestionModel{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeletePendingByUserID deletes all pending AI suggestions for a given user.
func (r *aiSuggestionRepository) DeletePendingByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	result := r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND status = ?", userID, string(entity.SuggestionStatusPending)).
		Delete(&model.AISuggestionModel{})
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

// ExistsPendingByUserID checks if there are any pending suggestions for a user.
func (r *aiSuggestionRepository) ExistsPendingByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.AISuggestionModel{}).
		Where("user_id = ? AND status = ?", userID, string(entity.SuggestionStatusPending)).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

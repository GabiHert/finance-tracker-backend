// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

// transactionRepository implements the adapter.TransactionRepository interface.
type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository instance.
func NewTransactionRepository(db *gorm.DB) adapter.TransactionRepository {
	return &transactionRepository{
		db: db,
	}
}

// Create creates a new transaction in the database.
func (r *transactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	transactionModel := model.TransactionFromEntity(transaction)
	result := r.db.WithContext(ctx).Create(transactionModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID retrieves a transaction by its ID.
func (r *transactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	var transactionModel model.TransactionModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&transactionModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrTransactionNotFound
		}
		return nil, result.Error
	}
	return transactionModel.ToEntity(), nil
}

// FindByIDWithCategory retrieves a transaction with its category by ID.
func (r *transactionRepository) FindByIDWithCategory(ctx context.Context, id uuid.UUID) (*entity.TransactionWithCategory, error) {
	var transactionModel model.TransactionModel
	result := r.db.WithContext(ctx).
		Preload("Category").
		Where("id = ?", id).
		First(&transactionModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrTransactionNotFound
		}
		return nil, result.Error
	}
	return transactionModel.ToEntityWithCategory(), nil
}

// FindByUser retrieves all transactions for a given user.
func (r *transactionRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	var transactionModels []model.TransactionModel
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("date DESC, created_at DESC").
		Find(&transactionModels)
	if result.Error != nil {
		return nil, result.Error
	}

	transactions := make([]*entity.Transaction, len(transactionModels))
	for i, tm := range transactionModels {
		transactions[i] = tm.ToEntity()
	}
	return transactions, nil
}

// FindByFilter retrieves transactions based on filter criteria with pagination.
func (r *transactionRepository) FindByFilter(ctx context.Context, filter adapter.TransactionFilter, pagination adapter.TransactionPagination) (*adapter.TransactionListResult, error) {
	query := r.db.WithContext(ctx).Model(&model.TransactionModel{})

	// Apply filters
	query = query.Where("user_id = ?", filter.UserID)

	if filter.StartDate != nil {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("date <= ?", filter.EndDate)
	}
	if len(filter.CategoryIDs) > 0 {
		query = query.Where("category_id IN ?", filter.CategoryIDs)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", string(*filter.Type))
	}
	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(description) LIKE ?", searchPattern)
	}

	// Get total count
	var total int64
	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (pagination.Page - 1) * pagination.Limit
	totalPages := int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))
	if totalPages == 0 {
		totalPages = 1
	}

	// Fetch transactions with category preloaded
	var transactionModels []model.TransactionModel
	result := query.
		Preload("Category").
		Order("date DESC, created_at DESC").
		Offset(offset).
		Limit(pagination.Limit).
		Find(&transactionModels)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to entities
	transactions := make([]*entity.TransactionWithCategory, len(transactionModels))
	for i, tm := range transactionModels {
		transactions[i] = tm.ToEntityWithCategory()
	}

	return &adapter.TransactionListResult{
		Transactions: transactions,
		Total:        total,
		Page:         pagination.Page,
		Limit:        pagination.Limit,
		TotalPages:   totalPages,
	}, nil
}

// GetTotals calculates totals for transactions based on filter criteria.
func (r *transactionRepository) GetTotals(ctx context.Context, filter adapter.TransactionFilter) (*adapter.TransactionTotals, error) {
	query := r.db.WithContext(ctx).Model(&model.TransactionModel{})

	// Apply filters
	query = query.Where("user_id = ?", filter.UserID)

	if filter.StartDate != nil {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("date <= ?", filter.EndDate)
	}
	if len(filter.CategoryIDs) > 0 {
		query = query.Where("category_id IN ?", filter.CategoryIDs)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", string(*filter.Type))
	}
	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(description) LIKE ?", searchPattern)
	}

	// Calculate income total
	var incomeTotal decimal.Decimal
	incomeQuery := query.Session(&gorm.Session{}).Where("type = ?", string(entity.TransactionTypeIncome))
	var incomeResult struct {
		Total decimal.Decimal
	}
	incomeQuery.Select("COALESCE(SUM(amount), 0) as total").Scan(&incomeResult)
	incomeTotal = incomeResult.Total

	// Calculate expense total
	var expenseTotal decimal.Decimal
	expenseQuery := query.Session(&gorm.Session{}).Where("type = ?", string(entity.TransactionTypeExpense))
	var expenseResult struct {
		Total decimal.Decimal
	}
	expenseQuery.Select("COALESCE(SUM(amount), 0) as total").Scan(&expenseResult)
	expenseTotal = expenseResult.Total

	// Calculate net total
	netTotal := incomeTotal.Add(expenseTotal)

	return &adapter.TransactionTotals{
		IncomeTotal:  incomeTotal,
		ExpenseTotal: expenseTotal,
		NetTotal:     netTotal,
	}, nil
}

// Update updates an existing transaction in the database.
func (r *transactionRepository) Update(ctx context.Context, transaction *entity.Transaction) error {
	transactionModel := model.TransactionFromEntity(transaction)
	result := r.db.WithContext(ctx).Save(transactionModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete soft-deletes a transaction from the database.
func (r *transactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.TransactionModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// BulkDelete soft-deletes multiple transactions by their IDs.
func (r *transactionRepository) BulkDelete(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error) {
	// Use transaction to ensure atomicity
	var deletedCount int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.TransactionModel{})
		if result.Error != nil {
			return result.Error
		}
		deletedCount = result.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return deletedCount, nil
}

// BulkUpdateCategory updates the category for multiple transactions.
func (r *transactionRepository) BulkUpdateCategory(ctx context.Context, ids []uuid.UUID, categoryID uuid.UUID, userID uuid.UUID) (int64, error) {
	// Use transaction to ensure atomicity
	var updatedCount int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.TransactionModel{}).
			Where("id IN ? AND user_id = ?", ids, userID).
			Updates(map[string]interface{}{
				"category_id": categoryID,
				"updated_at":  time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		updatedCount = result.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return updatedCount, nil
}

// ExistsByIDAndUser checks if a transaction exists for a given ID and user.
func (r *transactionRepository) ExistsByIDAndUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("id = ? AND user_id = ?", id, userID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// ExistsAllByIDsAndUser checks if all transactions exist for the given IDs and user.
func (r *transactionRepository) ExistsAllByIDsAndUser(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("id IN ? AND user_id = ?", ids, userID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count == int64(len(ids)), nil
}

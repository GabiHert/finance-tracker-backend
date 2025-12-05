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
	expandedBillIDs := []uuid.UUID{}

	for i, tm := range transactionModels {
		transactions[i] = tm.ToEntityWithCategory()
		// Collect IDs of expanded bills to count linked transactions
		if tm.ExpandedAt != nil {
			expandedBillIDs = append(expandedBillIDs, tm.ID)
		}
	}

	// Count linked transactions for expanded bills
	if len(expandedBillIDs) > 0 {
		var linkedCounts []struct {
			BillID uuid.UUID `gorm:"column:credit_card_payment_id"`
			Count  int       `gorm:"column:count"`
		}
		if err := r.db.WithContext(ctx).Model(&model.TransactionModel{}).
			Select("credit_card_payment_id, COUNT(*) as count").
			Where("credit_card_payment_id IN ?", expandedBillIDs).
			Where("is_hidden = ?", false).
			Group("credit_card_payment_id").
			Find(&linkedCounts).Error; err == nil {
			// Build map for fast lookup
			countMap := make(map[uuid.UUID]int)
			for _, lc := range linkedCounts {
				countMap[lc.BillID] = lc.Count
			}
			// Update transactions with linked count
			for _, txn := range transactions {
				if txn.Transaction.ExpandedAt != nil {
					txn.LinkedTransactionCount = countMap[txn.Transaction.ID]
				}
			}
		}
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

// BulkUpdateCategoryByPattern updates category for uncategorized transactions matching pattern.
func (r *transactionRepository) BulkUpdateCategoryByPattern(
	ctx context.Context,
	pattern string,
	categoryID uuid.UUID,
	ownerType entity.OwnerType,
	ownerID uuid.UUID,
) (int, error) {
	var result *gorm.DB
	now := time.Now().UTC()

	if ownerType == entity.OwnerTypeUser {
		// For user: update transactions belonging to user that have no category
		result = r.db.WithContext(ctx).
			Model(&model.TransactionModel{}).
			Where("user_id = ?", ownerID).
			Where("category_id IS NULL").
			Where("description ~* ?", pattern).
			Updates(map[string]interface{}{
				"category_id": categoryID,
				"updated_at":  now,
			})
	} else {
		// For group: update transactions belonging to any group member that have no category
		result = r.db.WithContext(ctx).
			Model(&model.TransactionModel{}).
			Where("user_id IN (SELECT user_id FROM group_members WHERE group_id = ? AND deleted_at IS NULL)", ownerID).
			Where("category_id IS NULL").
			Where("description ~* ?", pattern).
			Updates(map[string]interface{}{
				"category_id": categoryID,
				"updated_at":  now,
			})
	}

	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
}

// Credit card import methods

// FindPotentialBillPayments finds potential bill payment matches for CC import.
// It searches for transactions matching "Pagamento de fatura" or similar patterns
// within the specified date range.
func (r *transactionRepository) FindPotentialBillPayments(
	ctx context.Context,
	userID uuid.UUID,
	startDate time.Time,
	endDate time.Time,
) ([]*entity.Transaction, error) {
	var transactionModels []model.TransactionModel

	// Search for bill payment patterns - typically "Pagamento de fatura" or similar
	// These are expense transactions (negative amounts in the bank statement)
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Where("type = ?", string(entity.TransactionTypeExpense)).
		Where("is_credit_card_payment = ? OR description ~* ?", true, "pagamento.*fatura|fatura.*cartao|cartao.*credito").
		Where("expanded_at IS NULL"). // Exclude already expanded bills
		Order("date DESC").
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

// GetLinkedTransactions retrieves all CC transactions linked to a bill payment.
func (r *transactionRepository) GetLinkedTransactions(
	ctx context.Context,
	billPaymentID uuid.UUID,
) ([]*entity.Transaction, error) {
	var transactionModels []model.TransactionModel

	result := r.db.WithContext(ctx).
		Where("credit_card_payment_id = ?", billPaymentID).
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

// BulkCreateCCTransactions creates multiple CC transactions in a single operation.
// It also updates the bill payment (zeroing amount, setting expanded_at, billing_cycle, etc.).
func (r *transactionRepository) BulkCreateCCTransactions(
	ctx context.Context,
	transactions []*entity.Transaction,
	billPaymentID uuid.UUID,
	originalAmount decimal.Decimal,
	billingCycle string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		// Create all CC transactions
		for _, txn := range transactions {
			transactionModel := model.TransactionFromEntity(txn)
			if err := tx.Create(transactionModel).Error; err != nil {
				return err
			}
		}

		// Update the bill payment: zero amount, set original_amount, set expanded_at, set billing_cycle
		result := tx.Model(&model.TransactionModel{}).
			Where("id = ?", billPaymentID).
			Updates(map[string]interface{}{
				"original_amount":        originalAmount,
				"amount":                 decimal.Zero,
				"expanded_at":            now,
				"is_credit_card_payment": true,
				"billing_cycle":          billingCycle,
				"updated_at":             now,
			})

		if result.Error != nil {
			return result.Error
		}

		return nil
	})
}

// BulkCreateStandaloneCCTransactions creates CC transactions without linking to a bill payment.
// Used when importing CC transactions without a matching bill.
func (r *transactionRepository) BulkCreateStandaloneCCTransactions(
	ctx context.Context,
	transactions []*entity.Transaction,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create all CC transactions without updating any bill payment
		for _, txn := range transactions {
			transactionModel := model.TransactionFromEntity(txn)
			if err := tx.Create(transactionModel).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// ExpandBillPayment marks a bill payment as expanded and zeroes its amount.
func (r *transactionRepository) ExpandBillPayment(
	ctx context.Context,
	billPaymentID uuid.UUID,
	originalAmount decimal.Decimal,
) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("id = ?", billPaymentID).
		Updates(map[string]interface{}{
			"original_amount":       originalAmount,
			"amount":                decimal.Zero,
			"expanded_at":           now,
			"is_credit_card_payment": true,
			"updated_at":            now,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// CollapseExpansion deletes all linked CC transactions and restores the bill payment.
func (r *transactionRepository) CollapseExpansion(ctx context.Context, billPaymentID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		// First, get the original amount from the bill payment
		var billPayment model.TransactionModel
		if err := tx.Where("id = ?", billPaymentID).First(&billPayment).Error; err != nil {
			return err
		}

		// Delete all linked CC transactions (hard delete)
		if err := tx.Unscoped().
			Where("credit_card_payment_id = ?", billPaymentID).
			Delete(&model.TransactionModel{}).Error; err != nil {
			return err
		}

		// Restore the bill payment: restore original_amount to amount, clear expanded_at and billing_cycle
		updates := map[string]interface{}{
			"expanded_at":   nil,
			"billing_cycle": "",
			"updated_at":    now,
		}

		// If original_amount is set, restore it to amount
		if billPayment.OriginalAmount != nil {
			updates["amount"] = *billPayment.OriginalAmount
			updates["original_amount"] = nil
		}

		result := tx.Model(&model.TransactionModel{}).
			Where("id = ?", billPaymentID).
			Updates(updates)

		return result.Error
	})
}

// GetCreditCardStatus retrieves the CC status for a specific billing cycle.
func (r *transactionRepository) GetCreditCardStatus(
	ctx context.Context,
	userID uuid.UUID,
	billingCycle string,
) (*adapter.CreditCardStatus, error) {
	// Find bill payment for this billing cycle (by billing_cycle field or date range)
	var billPayment model.TransactionModel

	// First try to find by explicit billing_cycle
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("is_credit_card_payment = ?", true).
		Where("billing_cycle = ?", billingCycle).
		First(&billPayment)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// If not found by billing_cycle, try to find by expanded CC transactions
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Look for any expanded bill with linked transactions in this billing cycle
		var linkedTxn model.TransactionModel
		result = r.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Where("billing_cycle = ?", billingCycle).
			Where("credit_card_payment_id IS NOT NULL").
			First(&linkedTxn)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Check for standalone CC transactions (no linked bill payment)
				var standaloneTxns []model.TransactionModel
				standaloneResult := r.db.WithContext(ctx).
					Where("user_id = ?", userID).
					Where("billing_cycle = ?", billingCycle).
					Where("credit_card_payment_id IS NULL").
					Where("deleted_at IS NULL").
					Where("is_hidden = ?", false).
					Find(&standaloneTxns)

				if standaloneResult.Error != nil {
					return nil, standaloneResult.Error
				}

				if len(standaloneTxns) > 0 {
					// Calculate total spending from standalone transactions
					totalSpending := decimal.Zero
					var transactions []*entity.Transaction
					for _, txn := range standaloneTxns {
						transactions = append(transactions, txn.ToEntity())
						totalSpending = totalSpending.Add(txn.Amount.Abs())
					}

					return &adapter.CreditCardStatus{
						BillingCycle:       billingCycle,
						IsExpanded:         false,
						LinkedTransactions: transactions,
						OriginalAmount:     &totalSpending,
						CurrentAmount:      &totalSpending,
					}, nil
				}

				// No CC data for this billing cycle
				return &adapter.CreditCardStatus{
					BillingCycle:       billingCycle,
					IsExpanded:         false,
					LinkedTransactions: []*entity.Transaction{},
				}, nil
			}
			return nil, result.Error
		}

		// Found linked transaction, get the bill payment
		if linkedTxn.CreditCardPaymentID != nil {
			if err := r.db.WithContext(ctx).
				Where("id = ?", *linkedTxn.CreditCardPaymentID).
				First(&billPayment).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Bill payment was deleted but linked transactions still exist
					// Return these as orphaned CC transactions
					linkedTxns, linkedErr := r.GetLinkedTransactions(ctx, *linkedTxn.CreditCardPaymentID)
					if linkedErr != nil {
						// If we can't get linked transactions, return empty status
						return &adapter.CreditCardStatus{
							BillingCycle:       billingCycle,
							IsExpanded:         false,
							LinkedTransactions: []*entity.Transaction{},
						}, nil
					}
					// Return orphaned CC transactions without bill payment info
					return &adapter.CreditCardStatus{
						BillingCycle:       billingCycle,
						IsExpanded:         true, // Transactions exist, just orphaned
						LinkedTransactions: linkedTxns,
					}, nil
				}
				return nil, err
			}
		}
	}

	// Get linked transactions
	linkedTransactions, err := r.GetLinkedTransactions(ctx, billPayment.ID)
	if err != nil {
		return nil, err
	}

	status := &adapter.CreditCardStatus{
		BillingCycle:       billingCycle,
		IsExpanded:         billPayment.ExpandedAt != nil,
		BillPaymentID:      &billPayment.ID,
		BillPaymentDate:    &billPayment.Date,
		CurrentAmount:      &billPayment.Amount,
		LinkedTransactions: linkedTransactions,
		ExpandedAt:         billPayment.ExpandedAt,
	}

	if billPayment.OriginalAmount != nil {
		status.OriginalAmount = billPayment.OriginalAmount
	}

	return status, nil
}

// IsBillExpanded checks if a bill payment has been expanded.
func (r *transactionRepository) IsBillExpanded(ctx context.Context, billPaymentID uuid.UUID) (bool, error) {
	var billPayment model.TransactionModel

	result := r.db.WithContext(ctx).
		Select("expanded_at").
		Where("id = ?", billPaymentID).
		First(&billPayment)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, domainerror.ErrBillPaymentNotFound
		}
		return false, result.Error
	}

	return billPayment.ExpandedAt != nil, nil
}

// FindBillPaymentByID retrieves a bill payment transaction by ID with ownership check.
func (r *transactionRepository) FindBillPaymentByID(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) (*entity.Transaction, error) {
	var transactionModel model.TransactionModel

	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&transactionModel)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrBillPaymentNotFound
		}
		return nil, result.Error
	}

	return transactionModel.ToEntity(), nil
}

// FindMostRecentCCBillingCycle finds the most recent billing cycle with CC transactions.
// Returns empty string if no CC transactions exist.
func (r *transactionRepository) FindMostRecentCCBillingCycle(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	var billingCycle string

	// Find the most recent billing cycle from transactions that have a billing_cycle set
	result := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Select("billing_cycle").
		Where("user_id = ?", userID).
		Where("billing_cycle IS NOT NULL").
		Where("billing_cycle != ''").
		Where("deleted_at IS NULL").
		Order("billing_cycle DESC").
		Limit(1).
		Scan(&billingCycle)

	if result.Error != nil {
		return "", result.Error
	}

	return billingCycle, nil
}

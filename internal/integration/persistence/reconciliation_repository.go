// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

// reconciliationRepository implements the adapter.ReconciliationRepository interface.
type reconciliationRepository struct {
	db *gorm.DB
}

// NewReconciliationRepository creates a new reconciliation repository instance.
func NewReconciliationRepository(db *gorm.DB) adapter.ReconciliationRepository {
	return &reconciliationRepository{
		db: db,
	}
}

// GetPendingBillingCycles retrieves billing cycles with unlinked CC transactions.
func (r *reconciliationRepository) GetPendingBillingCycles(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]adapter.PendingCycleData, error) {
	var results []struct {
		BillingCycle     string
		TransactionCount int
		TotalAmount      decimal.Decimal
		OldestDate       time.Time
		NewestDate       time.Time
	}

	// Query for billing cycles that have CC transactions but are not linked to any bill
	// (credit_card_payment_id IS NULL means not linked)
	// Also exclude expanded bill payments (expanded_at IS NOT NULL) as they are the parent bills, not CC transactions
	err := r.db.WithContext(ctx).
		Table("transactions").
		Select(`
			billing_cycle,
			COUNT(*) as transaction_count,
			ABS(SUM(amount)) as total_amount,
			MIN(date) as oldest_date,
			MAX(date) as newest_date
		`).
		Where("user_id = ?", userID).
		Where("billing_cycle IS NOT NULL").
		Where("billing_cycle != ''").
		Where("credit_card_payment_id IS NULL").  // Not linked to any bill
		Where("expanded_at IS NULL").             // Not an expanded bill payment
		Where("is_hidden = ?", false).
		Where("deleted_at IS NULL").
		Group("billing_cycle").
		Order("billing_cycle DESC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	cycles := make([]adapter.PendingCycleData, len(results))
	for i, r := range results {
		cycles[i] = adapter.PendingCycleData{
			BillingCycle:     r.BillingCycle,
			TransactionCount: r.TransactionCount,
			TotalAmount:      r.TotalAmount.IntPart(),
			OldestDate:       r.OldestDate,
			NewestDate:       r.NewestDate,
		}
	}

	return cycles, nil
}

// GetLinkedBillingCycles retrieves billing cycles with linked bill payments.
func (r *reconciliationRepository) GetLinkedBillingCycles(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]adapter.LinkedCycleData, error) {
	var results []struct {
		BillingCycle     string
		TransactionCount int
		TotalAmount      decimal.Decimal
		BillID           uuid.UUID
		BillDate         time.Time
		BillDescription  string
		BillAmount       decimal.Decimal
		CategoryName     *string
	}

	// Query for billing cycles that have CC transactions linked to a bill
	// Join with the bill payment transaction to get its details
	err := r.db.WithContext(ctx).
		Table("transactions t").
		Select(`
			t.billing_cycle,
			COUNT(*) as transaction_count,
			ABS(SUM(t.amount)) as total_amount,
			b.id as bill_id,
			b.date as bill_date,
			b.description as bill_description,
			ABS(COALESCE(b.original_amount, b.amount)) as bill_amount,
			c.name as category_name
		`).
		Joins("INNER JOIN transactions b ON t.credit_card_payment_id = b.id AND b.deleted_at IS NULL").
		Joins("LEFT JOIN categories c ON b.category_id = c.id AND c.deleted_at IS NULL").
		Where("t.user_id = ?", userID).
		Where("t.billing_cycle IS NOT NULL").
		Where("t.billing_cycle != ''").
		Where("t.credit_card_payment_id IS NOT NULL"). // Linked to a bill
		Where("t.is_hidden = ?", false).
		Where("t.deleted_at IS NULL").
		Group("t.billing_cycle, b.id, b.date, b.description, b.original_amount, b.amount, c.name").
		Order("t.billing_cycle DESC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	cycles := make([]adapter.LinkedCycleData, len(results))
	for i, r := range results {
		cycles[i] = adapter.LinkedCycleData{
			BillingCycle:     r.BillingCycle,
			TransactionCount: r.TransactionCount,
			TotalAmount:      r.TotalAmount.IntPart(),
			BillID:           r.BillID,
			BillDate:         r.BillDate,
			BillDescription:  r.BillDescription,
			BillAmount:       r.BillAmount.IntPart(),
			CategoryName:     r.CategoryName,
		}
	}

	return cycles, nil
}

// FindPotentialBills finds bill payments that could match a billing cycle.
func (r *reconciliationRepository) FindPotentialBills(
	ctx context.Context,
	userID uuid.UUID,
	billingCycle string,
	ccTotal decimal.Decimal,
	dateRange adapter.DateRange,
) ([]adapter.BillData, error) {
	var results []struct {
		ID           uuid.UUID
		Date         time.Time
		Description  string
		Amount       decimal.Decimal
		CategoryName *string
	}

	// Find bill payments (expense transactions that look like CC bill payments)
	// that are not already expanded and within the date range
	err := r.db.WithContext(ctx).
		Table("transactions t").
		Select(`
			t.id,
			t.date,
			t.description,
			ABS(t.amount) as amount,
			c.name as category_name
		`).
		Joins("LEFT JOIN categories c ON t.category_id = c.id AND c.deleted_at IS NULL").
		Where("t.user_id = ?", userID).
		Where("t.type = ?", string(entity.TransactionTypeExpense)).
		Where("t.date >= ? AND t.date <= ?", dateRange.Start, dateRange.End).
		Where("t.expanded_at IS NULL").             // Not already expanded
		Where("t.billing_cycle IS NULL OR t.billing_cycle = ''"). // Not a CC transaction
		Where("t.deleted_at IS NULL").
		Where("t.is_credit_card_payment = ? OR t.description ~* ?", true, "pagamento.*fatura|fatura.*cartao|cartao.*credito").
		Order("t.date DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	bills := make([]adapter.BillData, len(results))
	for i, r := range results {
		bills[i] = adapter.BillData{
			ID:           r.ID,
			Date:         r.Date,
			Description:  r.Description,
			Amount:       r.Amount.IntPart(),
			CategoryName: r.CategoryName,
		}
	}

	return bills, nil
}

// GetCCTransactionsByBillingCycle retrieves all CC transaction IDs for a billing cycle.
func (r *reconciliationRepository) GetCCTransactionsByBillingCycle(
	ctx context.Context,
	userID uuid.UUID,
	billingCycle string,
) ([]uuid.UUID, error) {
	var ids []uuid.UUID

	err := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Select("id").
		Where("user_id = ?", userID).
		Where("billing_cycle = ?", billingCycle).
		Where("deleted_at IS NULL").
		Scan(&ids).Error

	if err != nil {
		return nil, err
	}

	return ids, nil
}

// GetCCTotalByBillingCycle calculates the total amount for CC transactions in a billing cycle.
func (r *reconciliationRepository) GetCCTotalByBillingCycle(
	ctx context.Context,
	userID uuid.UUID,
	billingCycle string,
) (decimal.Decimal, int, error) {
	var result struct {
		Total decimal.Decimal
		Count int
	}

	err := r.db.WithContext(ctx).
		Table("transactions").
		Select("ABS(SUM(amount)) as total, COUNT(*) as count").
		Where("user_id = ?", userID).
		Where("billing_cycle = ?", billingCycle).
		Where("is_hidden = ?", false).
		Where("credit_card_payment_id IS NULL"). // Only pending (not linked)
		Where("deleted_at IS NULL").
		Scan(&result).Error

	if err != nil {
		return decimal.Zero, 0, err
	}

	return result.Total, result.Count, nil
}

// LinkCCTransactionsToBill links CC transactions to a bill payment.
func (r *reconciliationRepository) LinkCCTransactionsToBill(
	ctx context.Context,
	userID uuid.UUID,
	billingCycle string,
	billPaymentID uuid.UUID,
	originalBillAmount decimal.Decimal,
) (int, error) {
	var linkedCount int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		// Update all CC transactions for this billing cycle to link to the bill
		result := tx.Model(&model.TransactionModel{}).
			Where("user_id = ?", userID).
			Where("billing_cycle = ?", billingCycle).
			Where("credit_card_payment_id IS NULL"). // Only unlinked transactions
			Where("deleted_at IS NULL").
			Updates(map[string]interface{}{
				"credit_card_payment_id": billPaymentID,
				"updated_at":             now,
			})

		if result.Error != nil {
			return result.Error
		}
		linkedCount = result.RowsAffected

		// Update the bill payment: set original_amount, zero amount, set expanded_at
		billUpdate := tx.Model(&model.TransactionModel{}).
			Where("id = ?", billPaymentID).
			Updates(map[string]interface{}{
				"original_amount":        originalBillAmount,
				"amount":                 decimal.Zero,
				"expanded_at":            now,
				"is_credit_card_payment": true,
				"billing_cycle":          billingCycle,
				"updated_at":             now,
			})

		return billUpdate.Error
	})

	if err != nil {
		return 0, err
	}

	return int(linkedCount), nil
}

// UnlinkCCTransactionsFromBill unlinks CC transactions from their bill payment.
func (r *reconciliationRepository) UnlinkCCTransactionsFromBill(
	ctx context.Context,
	userID uuid.UUID,
	billingCycle string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		// First, get the bill payment ID that was linked
		var billPaymentID *uuid.UUID
		err := tx.Model(&model.TransactionModel{}).
			Select("credit_card_payment_id").
			Where("user_id = ?", userID).
			Where("billing_cycle = ?", billingCycle).
			Where("credit_card_payment_id IS NOT NULL").
			Where("deleted_at IS NULL").
			Limit(1).
			Scan(&billPaymentID).Error

		if err != nil {
			return err
		}

		// Unlink all CC transactions for this billing cycle
		result := tx.Model(&model.TransactionModel{}).
			Where("user_id = ?", userID).
			Where("billing_cycle = ?", billingCycle).
			Where("deleted_at IS NULL").
			Updates(map[string]interface{}{
				"credit_card_payment_id": nil,
				"updated_at":             now,
			})

		if result.Error != nil {
			return result.Error
		}

		// If there was a bill payment linked, restore it
		if billPaymentID != nil {
			// Get the bill payment to restore its original amount
			var billPayment model.TransactionModel
			if err := tx.Where("id = ?", *billPaymentID).First(&billPayment).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				// Bill was deleted, nothing to restore
				return nil
			}

			// Restore the bill payment
			updates := map[string]interface{}{
				"expanded_at":   nil,
				"billing_cycle": "",
				"updated_at":    now,
			}

			// Restore original amount if it was set
			if billPayment.OriginalAmount != nil {
				updates["amount"] = *billPayment.OriginalAmount
				updates["original_amount"] = nil
			}

			return tx.Model(&model.TransactionModel{}).
				Where("id = ?", *billPaymentID).
				Updates(updates).Error
		}

		return nil
	})
}

// IsBillLinked checks if a bill payment is already linked to CC transactions.
func (r *reconciliationRepository) IsBillLinked(ctx context.Context, billID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("credit_card_payment_id = ?", billID).
		Where("deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsCycleLinked checks if a billing cycle already has a linked bill payment.
func (r *reconciliationRepository) IsCycleLinked(
	ctx context.Context,
	userID uuid.UUID,
	billingCycle string,
) (bool, *uuid.UUID, error) {
	var billID *uuid.UUID

	err := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Select("credit_card_payment_id").
		Where("user_id = ?", userID).
		Where("billing_cycle = ?", billingCycle).
		Where("credit_card_payment_id IS NOT NULL").
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(&billID).Error

	if err != nil {
		return false, nil, err
	}

	return billID != nil, billID, nil
}

// GetReconciliationSummary retrieves summary statistics for reconciliation.
func (r *reconciliationRepository) GetReconciliationSummary(
	ctx context.Context,
	userID uuid.UUID,
) (*adapter.ReconciliationSummaryData, error) {
	var pendingCount int
	var linkedCount int
	var totalCycles int

	// Count pending cycles (billing cycles with unlinked CC transactions)
	// Exclude expanded bill payments
	err := r.db.WithContext(ctx).
		Table("transactions").
		Select("COUNT(DISTINCT billing_cycle)").
		Where("user_id = ?", userID).
		Where("billing_cycle IS NOT NULL").
		Where("billing_cycle != ''").
		Where("credit_card_payment_id IS NULL").
		Where("expanded_at IS NULL").
		Where("is_hidden = ?", false).
		Where("deleted_at IS NULL").
		Scan(&pendingCount).Error

	if err != nil {
		return nil, err
	}

	// Count linked cycles (billing cycles with linked CC transactions)
	err = r.db.WithContext(ctx).
		Table("transactions").
		Select("COUNT(DISTINCT billing_cycle)").
		Where("user_id = ?", userID).
		Where("billing_cycle IS NOT NULL").
		Where("billing_cycle != ''").
		Where("credit_card_payment_id IS NOT NULL").
		Where("deleted_at IS NULL").
		Scan(&linkedCount).Error

	if err != nil {
		return nil, err
	}

	// Count total distinct billing cycles
	err = r.db.WithContext(ctx).
		Table("transactions").
		Select("COUNT(DISTINCT billing_cycle)").
		Where("user_id = ?", userID).
		Where("billing_cycle IS NOT NULL").
		Where("billing_cycle != ''").
		Where("deleted_at IS NULL").
		Scan(&totalCycles).Error

	if err != nil {
		return nil, err
	}

	return &adapter.ReconciliationSummaryData{
		TotalPending:  pendingCount,
		TotalLinked:   linkedCount,
		MonthsCovered: totalCycles,
	}, nil
}

// GetBillPaymentByID retrieves a bill payment by ID with ownership verification.
func (r *reconciliationRepository) GetBillPaymentByID(
	ctx context.Context,
	billID uuid.UUID,
	userID uuid.UUID,
) (*adapter.BillData, error) {
	var result struct {
		ID           uuid.UUID
		Date         time.Time
		Description  string
		Amount       decimal.Decimal
		CategoryName *string
	}

	err := r.db.WithContext(ctx).
		Table("transactions t").
		Select(`
			t.id,
			t.date,
			t.description,
			ABS(t.amount) as amount,
			c.name as category_name
		`).
		Joins("LEFT JOIN categories c ON t.category_id = c.id AND c.deleted_at IS NULL").
		Where("t.id = ?", billID).
		Where("t.user_id = ?", userID).
		Where("t.deleted_at IS NULL").
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &adapter.BillData{
		ID:           result.ID,
		Date:         result.Date,
		Description:  result.Description,
		Amount:       result.Amount.IntPart(),
		CategoryName: result.CategoryName,
	}, nil
}

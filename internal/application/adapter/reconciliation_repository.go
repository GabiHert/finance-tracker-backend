// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReconciliationRepository defines the interface for reconciliation persistence operations.
type ReconciliationRepository interface {
	// GetPendingBillingCycles retrieves billing cycles with unlinked CC transactions.
	// Returns cycles where CC transactions exist but no bill payment is linked.
	GetPendingBillingCycles(
		ctx context.Context,
		userID uuid.UUID,
		limit int,
		offset int,
	) ([]PendingCycleData, error)

	// GetLinkedBillingCycles retrieves billing cycles with linked bill payments.
	// Returns cycles where CC transactions are linked to a bill payment.
	GetLinkedBillingCycles(
		ctx context.Context,
		userID uuid.UUID,
		limit int,
		offset int,
	) ([]LinkedCycleData, error)

	// FindPotentialBills finds bill payments that could match a billing cycle.
	// Uses date range and amount tolerance to find potential matches.
	FindPotentialBills(
		ctx context.Context,
		userID uuid.UUID,
		billingCycle string,
		ccTotal decimal.Decimal,
		dateRange DateRange,
	) ([]BillData, error)

	// GetCCTransactionsByBillingCycle retrieves all CC transaction IDs for a billing cycle.
	GetCCTransactionsByBillingCycle(
		ctx context.Context,
		userID uuid.UUID,
		billingCycle string,
	) ([]uuid.UUID, error)

	// GetCCTotalByBillingCycle calculates the total amount for CC transactions in a billing cycle.
	GetCCTotalByBillingCycle(
		ctx context.Context,
		userID uuid.UUID,
		billingCycle string,
	) (decimal.Decimal, int, error)

	// LinkCCTransactionsToBill links CC transactions to a bill payment.
	// Updates all CC transactions in the billing cycle to reference the bill payment.
	LinkCCTransactionsToBill(
		ctx context.Context,
		userID uuid.UUID,
		billingCycle string,
		billPaymentID uuid.UUID,
		originalBillAmount decimal.Decimal,
	) (int, error)

	// UnlinkCCTransactionsFromBill unlinks CC transactions from their bill payment.
	// Sets credit_card_payment_id to NULL for all transactions in the billing cycle.
	UnlinkCCTransactionsFromBill(
		ctx context.Context,
		userID uuid.UUID,
		billingCycle string,
	) error

	// IsBillLinked checks if a bill payment is already linked to CC transactions.
	IsBillLinked(ctx context.Context, billID uuid.UUID) (bool, error)

	// IsCycleLinked checks if a billing cycle already has a linked bill payment.
	IsCycleLinked(ctx context.Context, userID uuid.UUID, billingCycle string) (bool, *uuid.UUID, error)

	// GetReconciliationSummary retrieves summary statistics for reconciliation.
	GetReconciliationSummary(
		ctx context.Context,
		userID uuid.UUID,
	) (*ReconciliationSummaryData, error)

	// GetBillPaymentByID retrieves a bill payment by ID with ownership verification.
	GetBillPaymentByID(
		ctx context.Context,
		billID uuid.UUID,
		userID uuid.UUID,
	) (*BillData, error)
}

// PendingCycleData represents a billing cycle with pending CC transactions.
type PendingCycleData struct {
	BillingCycle     string
	TransactionCount int
	TotalAmount      int64 // cents
	OldestDate       time.Time
	NewestDate       time.Time
}

// LinkedCycleData represents a billing cycle linked to a bill payment.
type LinkedCycleData struct {
	BillingCycle     string
	TransactionCount int
	TotalAmount      int64 // cents
	BillID           uuid.UUID
	BillDate         time.Time
	BillDescription  string
	BillAmount       int64 // cents (original_amount)
	CategoryName     *string
}

// BillData represents a potential bill payment for matching.
type BillData struct {
	ID           uuid.UUID
	Date         time.Time
	Description  string
	Amount       int64 // cents (absolute value)
	CategoryName *string
}

// DateRange represents a date range for bill matching.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// ReconciliationSummaryData contains summary statistics for reconciliation.
type ReconciliationSummaryData struct {
	TotalPending  int
	TotalLinked   int
	MonthsCovered int
}

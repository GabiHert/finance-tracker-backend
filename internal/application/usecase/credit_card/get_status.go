// Package creditcard contains credit card import-related use cases.
package creditcard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// GetStatusInput represents the input for getting CC status.
type GetStatusInput struct {
	UserID       uuid.UUID
	BillingCycle string // Format: "YYYY-MM", empty for current month
}

// CCTransactionSummary represents a summary of a CC transaction.
type CCTransactionSummary struct {
	ID          uuid.UUID
	Date        time.Time
	Description string
	Amount      decimal.Decimal
	CategoryID  *uuid.UUID
	IsHidden    bool
}

// GetStatusOutput represents the CC status for a billing cycle.
type GetStatusOutput struct {
	BillingCycle        string
	IsExpanded          bool
	BillPaymentID       *uuid.UUID
	BillPaymentDate     *time.Time
	OriginalAmount      *decimal.Decimal
	CurrentAmount       *decimal.Decimal
	LinkedTransactions  []CCTransactionSummary
	ExpandedAt          *time.Time
}

// GetStatusUseCase handles getting CC status logic.
type GetStatusUseCase struct {
	transactionRepo adapter.TransactionRepository
}

// NewGetStatusUseCase creates a new GetStatusUseCase instance.
func NewGetStatusUseCase(transactionRepo adapter.TransactionRepository) *GetStatusUseCase {
	return &GetStatusUseCase{
		transactionRepo: transactionRepo,
	}
}

// Execute retrieves the CC status for a billing cycle.
func (uc *GetStatusUseCase) Execute(ctx context.Context, input GetStatusInput) (*GetStatusOutput, error) {
	billingCycle := input.BillingCycle

	// If no billing cycle specified, find the most recent one with CC transactions
	if billingCycle == "" {
		recentCycle, err := uc.transactionRepo.FindMostRecentCCBillingCycle(ctx, input.UserID)
		if err != nil {
			return nil, err
		}

		// If found, use it; otherwise default to current month
		if recentCycle != "" {
			billingCycle = recentCycle
		} else {
			billingCycle = time.Now().Format("2006-01")
		}
	}

	// Validate billing cycle format
	if !billingCycleRegex.MatchString(billingCycle) {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeInvalidBillingCycle,
			"billing cycle must be in YYYY-MM format",
			domainerror.ErrInvalidBillingCycle,
		)
	}

	// Get CC status from repository
	status, err := uc.transactionRepo.GetCreditCardStatus(ctx, input.UserID, billingCycle)
	if err != nil {
		return nil, err
	}

	// Convert linked transactions to summaries
	var transactionSummaries []CCTransactionSummary
	for _, txn := range status.LinkedTransactions {
		transactionSummaries = append(transactionSummaries, CCTransactionSummary{
			ID:          txn.ID,
			Date:        txn.Date,
			Description: txn.Description,
			Amount:      txn.Amount,
			CategoryID:  txn.CategoryID,
			IsHidden:    txn.IsHidden,
		})
	}

	return &GetStatusOutput{
		BillingCycle:        billingCycle,
		IsExpanded:          status.IsExpanded,
		BillPaymentID:       status.BillPaymentID,
		BillPaymentDate:     status.BillPaymentDate,
		OriginalAmount:      status.OriginalAmount,
		CurrentAmount:       status.CurrentAmount,
		LinkedTransactions:  transactionSummaries,
		ExpandedAt:          status.ExpandedAt,
	}, nil
}

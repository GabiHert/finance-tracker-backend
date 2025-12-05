// Package transaction contains transaction-related use cases.
package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// ListTransactionsInput represents the input for listing transactions.
type ListTransactionsInput struct {
	UserID      uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	CategoryIDs []uuid.UUID
	Type        *entity.TransactionType
	Search      string
	GroupByDate bool
	Page        int
	Limit       int
}

// TransactionOutput represents a single transaction in the output.
type TransactionOutput struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Date        time.Time
	Description string
	Amount      decimal.Decimal
	Type        entity.TransactionType
	CategoryID  *uuid.UUID
	Category    *CategoryOutput
	Notes       string
	IsRecurring bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Credit card import fields
	BillingCycle           string // "YYYY-MM" format if imported from CC statement
	IsExpandedBill         bool   // True if this is a bill payment that has been expanded
	LinkedTransactionCount int    // Number of CC transactions linked to this bill
	InstallmentCurrent     *int   // Current installment number
	InstallmentTotal       *int   // Total installments
}

// CategoryOutput represents category information in transaction output.
type CategoryOutput struct {
	ID    uuid.UUID
	Name  string
	Color string
	Icon  string
	Type  entity.CategoryType
}

// PaginationOutput represents pagination information in the output.
type PaginationOutput struct {
	Page       int
	Limit      int
	Total      int64
	TotalPages int
}

// TotalsOutput represents aggregated totals in the output.
type TotalsOutput struct {
	IncomeTotal  decimal.Decimal
	ExpenseTotal decimal.Decimal
	NetTotal     decimal.Decimal
}

// ListTransactionsOutput represents the output of listing transactions.
type ListTransactionsOutput struct {
	Transactions []*TransactionOutput
	Pagination   PaginationOutput
	Totals       TotalsOutput
}

// ListTransactionsUseCase handles listing transactions logic.
type ListTransactionsUseCase struct {
	transactionRepo adapter.TransactionRepository
}

// NewListTransactionsUseCase creates a new ListTransactionsUseCase instance.
func NewListTransactionsUseCase(transactionRepo adapter.TransactionRepository) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{
		transactionRepo: transactionRepo,
	}
}

// Execute performs the transaction listing.
func (uc *ListTransactionsUseCase) Execute(ctx context.Context, input ListTransactionsInput) (*ListTransactionsOutput, error) {
	// Set default pagination values
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Build filter
	filter := adapter.TransactionFilter{
		UserID:      input.UserID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		CategoryIDs: input.CategoryIDs,
		Type:        input.Type,
		Search:      input.Search,
		GroupByDate: input.GroupByDate,
	}

	// Build pagination
	pagination := adapter.TransactionPagination{
		Page:  page,
		Limit: limit,
	}

	// Fetch transactions
	result, err := uc.transactionRepo.FindByFilter(ctx, filter, pagination)
	if err != nil {
		return nil, err
	}

	// Get totals
	totals, err := uc.transactionRepo.GetTotals(ctx, filter)
	if err != nil {
		// Log error but continue without totals
		totals = &adapter.TransactionTotals{}
	}

	// Build output
	output := &ListTransactionsOutput{
		Transactions: make([]*TransactionOutput, len(result.Transactions)),
		Pagination: PaginationOutput{
			Page:       result.Page,
			Limit:      result.Limit,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
		Totals: TotalsOutput{
			IncomeTotal:  totals.IncomeTotal,
			ExpenseTotal: totals.ExpenseTotal,
			NetTotal:     totals.NetTotal,
		},
	}

	for i, txnWithCat := range result.Transactions {
		txnOutput := &TransactionOutput{
			ID:                     txnWithCat.Transaction.ID,
			UserID:                 txnWithCat.Transaction.UserID,
			Date:                   txnWithCat.Transaction.Date,
			Description:            txnWithCat.Transaction.Description,
			Amount:                 txnWithCat.Transaction.Amount,
			Type:                   txnWithCat.Transaction.Type,
			CategoryID:             txnWithCat.Transaction.CategoryID,
			Notes:                  txnWithCat.Transaction.Notes,
			IsRecurring:            txnWithCat.Transaction.IsRecurring,
			CreatedAt:              txnWithCat.Transaction.CreatedAt,
			UpdatedAt:              txnWithCat.Transaction.UpdatedAt,
			BillingCycle:           txnWithCat.Transaction.BillingCycle,
			IsExpandedBill:         txnWithCat.Transaction.ExpandedAt != nil,
			LinkedTransactionCount: txnWithCat.LinkedTransactionCount,
			InstallmentCurrent:     txnWithCat.Transaction.InstallmentCurrent,
			InstallmentTotal:       txnWithCat.Transaction.InstallmentTotal,
		}

		// Add category if present
		if txnWithCat.Category != nil {
			txnOutput.Category = &CategoryOutput{
				ID:    txnWithCat.Category.ID,
				Name:  txnWithCat.Category.Name,
				Color: txnWithCat.Category.Color,
				Icon:  txnWithCat.Category.Icon,
				Type:  txnWithCat.Category.Type,
			}
		}

		output.Transactions[i] = txnOutput
	}

	return output, nil
}

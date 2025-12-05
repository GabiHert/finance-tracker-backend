// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"time"

	"github.com/finance-tracker/backend/internal/application/usecase/transaction"
)

// CreateTransactionRequest represents the request body for transaction creation.
type CreateTransactionRequest struct {
	Date        string  `json:"date" binding:"required"`
	Description string  `json:"description" binding:"required,min=1,max=255"`
	Amount      float64 `json:"amount" binding:"required"`
	Type        string  `json:"type" binding:"required,oneof=expense income"`
	CategoryID  *string `json:"category_id,omitempty"`
	Notes       string  `json:"notes,omitempty" binding:"omitempty,max=1000"`
	IsRecurring bool    `json:"is_recurring,omitempty"`
}

// UpdateTransactionRequest represents the request body for transaction update.
type UpdateTransactionRequest struct {
	Date          *string  `json:"date,omitempty"`
	Description   *string  `json:"description,omitempty" binding:"omitempty,min=1,max=255"`
	Amount        *float64 `json:"amount,omitempty"`
	Type          *string  `json:"type,omitempty" binding:"omitempty,oneof=expense income"`
	CategoryID    *string  `json:"category_id,omitempty"`
	ClearCategory bool     `json:"clear_category,omitempty"`
	Notes         *string  `json:"notes,omitempty" binding:"omitempty,max=1000"`
	IsRecurring   *bool    `json:"is_recurring,omitempty"`
}

// BulkDeleteTransactionsRequest represents the request body for bulk transaction deletion.
type BulkDeleteTransactionsRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

// BulkCategorizeTransactionsRequest represents the request body for bulk transaction categorization.
type BulkCategorizeTransactionsRequest struct {
	IDs        []string `json:"ids" binding:"required,min=1"`
	CategoryID string   `json:"category_id" binding:"required"`
}

// TransactionCategoryResponse represents category information in transaction response.
type TransactionCategoryResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
	Type  string `json:"type"`
}

// TransactionResponse represents a single transaction in API responses.
type TransactionResponse struct {
	ID          string                       `json:"id"`
	UserID      string                       `json:"user_id"`
	Date        string                       `json:"date"`
	Description string                       `json:"description"`
	Amount      string                       `json:"amount"`
	Type        string                       `json:"type"`
	CategoryID  *string                      `json:"category_id,omitempty"`
	Category    *TransactionCategoryResponse `json:"category,omitempty"`
	Notes       string                       `json:"notes"`
	IsRecurring bool                         `json:"is_recurring"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
	// Credit card import fields
	BillingCycle           string `json:"billing_cycle,omitempty"`
	IsExpandedBill         bool   `json:"is_expanded_bill,omitempty"`
	LinkedTransactionCount int    `json:"linked_transaction_count,omitempty"`
	InstallmentCurrent     *int   `json:"installment_current,omitempty"`
	InstallmentTotal       *int   `json:"installment_total,omitempty"`
}

// TransactionPaginationResponse represents pagination information in API responses.
type TransactionPaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// TransactionTotalsResponse represents aggregated totals in API responses.
type TransactionTotalsResponse struct {
	IncomeTotal  string `json:"income_total"`
	ExpenseTotal string `json:"expense_total"`
	NetTotal     string `json:"net_total"`
}

// TransactionListResponse represents the response for listing transactions.
type TransactionListResponse struct {
	Transactions []TransactionResponse         `json:"transactions"`
	Pagination   TransactionPaginationResponse `json:"pagination"`
	Totals       TransactionTotalsResponse     `json:"totals"`
}

// BulkDeleteTransactionsResponse represents the response for bulk transaction deletion.
type BulkDeleteTransactionsResponse struct {
	DeletedCount int64 `json:"deleted_count"`
}

// BulkCategorizeTransactionsResponse represents the response for bulk transaction categorization.
type BulkCategorizeTransactionsResponse struct {
	UpdatedCount int64 `json:"updated_count"`
}

// ToTransactionResponse converts a TransactionOutput to a TransactionResponse DTO.
func ToTransactionResponse(txn *transaction.TransactionOutput) TransactionResponse {
	response := TransactionResponse{
		ID:                     txn.ID.String(),
		UserID:                 txn.UserID.String(),
		Date:                   txn.Date.Format("2006-01-02"),
		Description:            txn.Description,
		Amount:                 txn.Amount.String(),
		Type:                   string(txn.Type),
		Notes:                  txn.Notes,
		IsRecurring:            txn.IsRecurring,
		CreatedAt:              txn.CreatedAt,
		UpdatedAt:              txn.UpdatedAt,
		BillingCycle:           txn.BillingCycle,
		IsExpandedBill:         txn.IsExpandedBill,
		LinkedTransactionCount: txn.LinkedTransactionCount,
		InstallmentCurrent:     txn.InstallmentCurrent,
		InstallmentTotal:       txn.InstallmentTotal,
	}

	if txn.CategoryID != nil {
		categoryIDStr := txn.CategoryID.String()
		response.CategoryID = &categoryIDStr
	}

	if txn.Category != nil {
		response.Category = &TransactionCategoryResponse{
			ID:    txn.Category.ID.String(),
			Name:  txn.Category.Name,
			Color: txn.Category.Color,
			Icon:  txn.Category.Icon,
			Type:  string(txn.Category.Type),
		}
	}

	return response
}

// ToTransactionListResponse converts a ListTransactionsOutput to TransactionListResponse.
func ToTransactionListResponse(output *transaction.ListTransactionsOutput) TransactionListResponse {
	transactions := make([]TransactionResponse, len(output.Transactions))
	for i, txn := range output.Transactions {
		transactions[i] = ToTransactionResponse(txn)
	}

	return TransactionListResponse{
		Transactions: transactions,
		Pagination: TransactionPaginationResponse{
			Page:       output.Pagination.Page,
			Limit:      output.Pagination.Limit,
			Total:      output.Pagination.Total,
			TotalPages: output.Pagination.TotalPages,
		},
		Totals: TransactionTotalsResponse{
			IncomeTotal:  output.Totals.IncomeTotal.String(),
			ExpenseTotal: output.Totals.ExpenseTotal.String(),
			NetTotal:     output.Totals.NetTotal.String(),
		},
	}
}

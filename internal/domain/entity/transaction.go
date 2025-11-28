// Package entity defines the core business entities for the domain layer.
package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionType represents the type of transaction (expense or income).
type TransactionType string

const (
	TransactionTypeExpense TransactionType = "expense"
	TransactionTypeIncome  TransactionType = "income"
)

// Transaction represents a financial transaction in the Finance Tracker system.
type Transaction struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Date        time.Time
	Description string
	Amount      decimal.Decimal // Negative for expenses, positive for income
	Type        TransactionType
	CategoryID  *uuid.UUID // Optional, can be uncategorized
	Notes       string
	IsRecurring bool
	UploadedAt  *time.Time // Timestamp when transaction was uploaded (for imports)
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // Soft-delete support
}

// NewTransaction creates a new Transaction entity.
func NewTransaction(
	userID uuid.UUID,
	date time.Time,
	description string,
	amount decimal.Decimal,
	transactionType TransactionType,
	categoryID *uuid.UUID,
	notes string,
	isRecurring bool,
) *Transaction {
	now := time.Now().UTC()

	return &Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		Date:        date,
		Description: description,
		Amount:      amount,
		Type:        transactionType,
		CategoryID:  categoryID,
		Notes:       notes,
		IsRecurring: isRecurring,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// TransactionWithCategory represents a transaction with its associated category.
type TransactionWithCategory struct {
	Transaction *Transaction
	Category    *Category
}

// TransactionListResult represents the result of listing transactions.
type TransactionListResult struct {
	Transactions []*TransactionWithCategory
	Total        int64
	Page         int
	Limit        int
	TotalPages   int
}

// TransactionTotals represents aggregated totals for transactions.
type TransactionTotals struct {
	IncomeTotal  decimal.Decimal
	ExpenseTotal decimal.Decimal
	NetTotal     decimal.Decimal
}

// TransactionsByDate represents transactions grouped by date.
type TransactionsByDate struct {
	Date         time.Time
	Transactions []*TransactionWithCategory
	DailyTotal   decimal.Decimal
}

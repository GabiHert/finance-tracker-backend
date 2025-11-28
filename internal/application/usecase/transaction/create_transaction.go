// Package transaction contains transaction-related use cases.
package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

const (
	// MaxDescriptionLength is the maximum allowed length for transaction descriptions.
	MaxDescriptionLength = 255
	// MaxNotesLength is the maximum allowed length for transaction notes.
	MaxNotesLength = 1000
)

// CreateTransactionInput represents the input for transaction creation.
type CreateTransactionInput struct {
	UserID      uuid.UUID
	Date        time.Time
	Description string
	Amount      decimal.Decimal
	Type        entity.TransactionType
	CategoryID  *uuid.UUID
	Notes       string
	IsRecurring bool
}

// CreateTransactionOutput represents the output of transaction creation.
type CreateTransactionOutput struct {
	Transaction *TransactionOutput
}

// CreateTransactionUseCase handles transaction creation logic.
type CreateTransactionUseCase struct {
	transactionRepo adapter.TransactionRepository
	categoryRepo    adapter.CategoryRepository
}

// NewCreateTransactionUseCase creates a new CreateTransactionUseCase instance.
func NewCreateTransactionUseCase(
	transactionRepo adapter.TransactionRepository,
	categoryRepo adapter.CategoryRepository,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

// Execute performs the transaction creation.
func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateTransactionInput) (*CreateTransactionOutput, error) {
	// Validate description length
	if len(input.Description) > MaxDescriptionLength {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeDescriptionTooLong,
			fmt.Sprintf("description must not exceed %d characters", MaxDescriptionLength),
			domainerror.ErrDescriptionTooLong,
		)
	}

	// Validate notes length
	if len(input.Notes) > MaxNotesLength {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeNotesTooLong,
			fmt.Sprintf("notes must not exceed %d characters", MaxNotesLength),
			domainerror.ErrNotesTooLong,
		)
	}

	// Validate transaction type
	if !isValidTransactionType(input.Type) {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeInvalidTransactionType,
			"transaction type must be 'expense' or 'income'",
			domainerror.ErrInvalidTransactionType,
		)
	}

	// Validate category if provided
	var category *entity.Category
	if input.CategoryID != nil {
		cat, err := uc.categoryRepo.FindByID(ctx, *input.CategoryID)
		if err != nil {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeTxnCategoryNotFound,
				"category not found",
				domainerror.ErrCategoryNotFoundForTransaction,
			)
		}

		// Verify category ownership
		if cat.OwnerType != entity.OwnerTypeUser || cat.OwnerID != input.UserID {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeTxnCategoryNotOwned,
				"category does not belong to user",
				domainerror.ErrCategoryNotOwnedByUser,
			)
		}

		category = cat
	}

	// Create transaction entity
	transaction := entity.NewTransaction(
		input.UserID,
		input.Date,
		input.Description,
		input.Amount,
		input.Type,
		input.CategoryID,
		input.Notes,
		input.IsRecurring,
	)

	// Save transaction to database
	if err := uc.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Build output
	output := &CreateTransactionOutput{
		Transaction: &TransactionOutput{
			ID:          transaction.ID,
			UserID:      transaction.UserID,
			Date:        transaction.Date,
			Description: transaction.Description,
			Amount:      transaction.Amount,
			Type:        transaction.Type,
			CategoryID:  transaction.CategoryID,
			Notes:       transaction.Notes,
			IsRecurring: transaction.IsRecurring,
			CreatedAt:   transaction.CreatedAt,
			UpdatedAt:   transaction.UpdatedAt,
		},
	}

	// Add category if present
	if category != nil {
		output.Transaction.Category = &CategoryOutput{
			ID:    category.ID,
			Name:  category.Name,
			Color: category.Color,
			Icon:  category.Icon,
			Type:  category.Type,
		}
	}

	return output, nil
}

// isValidTransactionType validates the transaction type.
func isValidTransactionType(transactionType entity.TransactionType) bool {
	return transactionType == entity.TransactionTypeExpense || transactionType == entity.TransactionTypeIncome
}

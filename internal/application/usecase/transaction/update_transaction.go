// Package transaction contains transaction-related use cases.
package transaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// UpdateTransactionInput represents the input for transaction update.
type UpdateTransactionInput struct {
	TransactionID uuid.UUID
	UserID        uuid.UUID
	Date          *time.Time
	Description   *string
	Amount        *decimal.Decimal
	Type          *entity.TransactionType
	CategoryID    *uuid.UUID
	ClearCategory bool // Set to true to remove category
	Notes         *string
	IsRecurring   *bool
}

// UpdateTransactionOutput represents the output of transaction update.
type UpdateTransactionOutput struct {
	Transaction *TransactionOutput
}

// UpdateTransactionUseCase handles transaction update logic.
type UpdateTransactionUseCase struct {
	transactionRepo adapter.TransactionRepository
	categoryRepo    adapter.CategoryRepository
}

// NewUpdateTransactionUseCase creates a new UpdateTransactionUseCase instance.
func NewUpdateTransactionUseCase(
	transactionRepo adapter.TransactionRepository,
	categoryRepo adapter.CategoryRepository,
) *UpdateTransactionUseCase {
	return &UpdateTransactionUseCase{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

// Execute performs the transaction update.
func (uc *UpdateTransactionUseCase) Execute(ctx context.Context, input UpdateTransactionInput) (*UpdateTransactionOutput, error) {
	// Find the existing transaction
	transaction, err := uc.transactionRepo.FindByID(ctx, input.TransactionID)
	if err != nil {
		if errors.Is(err, domainerror.ErrTransactionNotFound) {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeTransactionNotFound,
				"transaction not found",
				domainerror.ErrTransactionNotFound,
			)
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}

	// Check if user is authorized to update this transaction
	if transaction.UserID != input.UserID {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeNotAuthorizedTransaction,
			"not authorized to update this transaction",
			domainerror.ErrNotAuthorizedToModifyTransaction,
		)
	}

	// Update fields if provided
	if input.Date != nil {
		transaction.Date = *input.Date
	}

	if input.Description != nil {
		if len(*input.Description) > MaxDescriptionLength {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeDescriptionTooLong,
				fmt.Sprintf("description must not exceed %d characters", MaxDescriptionLength),
				domainerror.ErrDescriptionTooLong,
			)
		}
		transaction.Description = *input.Description
	}

	if input.Amount != nil {
		transaction.Amount = *input.Amount
	}

	if input.Type != nil {
		if !isValidTransactionType(*input.Type) {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeInvalidTransactionType,
				"transaction type must be 'expense' or 'income'",
				domainerror.ErrInvalidTransactionType,
			)
		}
		transaction.Type = *input.Type
	}

	// Handle category update
	var category *entity.Category
	if input.ClearCategory {
		transaction.CategoryID = nil
	} else if input.CategoryID != nil {
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

		transaction.CategoryID = input.CategoryID
		category = cat
	} else if transaction.CategoryID != nil {
		// Load existing category for response
		cat, err := uc.categoryRepo.FindByID(ctx, *transaction.CategoryID)
		if err == nil {
			category = cat
		}
	}

	if input.Notes != nil {
		if len(*input.Notes) > MaxNotesLength {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeNotesTooLong,
				fmt.Sprintf("notes must not exceed %d characters", MaxNotesLength),
				domainerror.ErrNotesTooLong,
			)
		}
		transaction.Notes = *input.Notes
	}

	if input.IsRecurring != nil {
		transaction.IsRecurring = *input.IsRecurring
	}

	// Update timestamp
	transaction.UpdatedAt = time.Now().UTC()

	// Save changes
	if err := uc.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// Build output
	output := &UpdateTransactionOutput{
		Transaction: &TransactionOutput{
			ID:                 transaction.ID,
			UserID:             transaction.UserID,
			Date:               transaction.Date,
			Description:        transaction.Description,
			Amount:             transaction.Amount,
			Type:               transaction.Type,
			CategoryID:         transaction.CategoryID,
			Notes:              transaction.Notes,
			IsRecurring:        transaction.IsRecurring,
			CreatedAt:          transaction.CreatedAt,
			UpdatedAt:          transaction.UpdatedAt,
			BillingCycle:       transaction.BillingCycle,
			IsExpandedBill:     transaction.ExpandedAt != nil,
			InstallmentCurrent: transaction.InstallmentCurrent,
			InstallmentTotal:   transaction.InstallmentTotal,
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

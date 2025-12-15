// Package transaction contains transaction-related use cases.
package transaction

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
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
	UserID              uuid.UUID
	Date                time.Time
	Description         string
	Amount              decimal.Decimal
	Type                entity.TransactionType
	CategoryID          *uuid.UUID
	Notes               string
	IsRecurring         bool
	BillingCycle        string // Format: "YYYY-MM" (e.g., "2024-11") - for credit card transactions
	IsCreditCardPayment bool   // True if this is a credit card payment transaction
}

// CreateTransactionOutput represents the output of transaction creation.
type CreateTransactionOutput struct {
	Transaction *TransactionOutput
}

// CreateTransactionUseCase handles transaction creation logic.
type CreateTransactionUseCase struct {
	transactionRepo  adapter.TransactionRepository
	categoryRepo     adapter.CategoryRepository
	categoryRuleRepo adapter.CategoryRuleRepository
}

// NewCreateTransactionUseCase creates a new CreateTransactionUseCase instance.
func NewCreateTransactionUseCase(
	transactionRepo adapter.TransactionRepository,
	categoryRepo adapter.CategoryRepository,
	categoryRuleRepo adapter.CategoryRuleRepository,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		transactionRepo:  transactionRepo,
		categoryRepo:     categoryRepo,
		categoryRuleRepo: categoryRuleRepo,
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

	// Validate category if provided, or auto-categorize if not
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
	} else {
		// Auto-categorize: try to match transaction description against category rules
		matchedCategoryID, matchedCategory := uc.autoCategorize(ctx, input.UserID, input.Description)
		if matchedCategoryID != nil {
			input.CategoryID = matchedCategoryID
			category = matchedCategory
		}
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

	// Set credit card fields if provided
	if input.BillingCycle != "" {
		transaction.BillingCycle = input.BillingCycle
	}
	if input.IsCreditCardPayment {
		transaction.IsCreditCardPayment = true
	}

	// Save transaction to database
	if err := uc.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Build output
	output := &CreateTransactionOutput{
		Transaction: &TransactionOutput{
			ID:                  transaction.ID,
			UserID:              transaction.UserID,
			Date:                transaction.Date,
			Description:         transaction.Description,
			Amount:              transaction.Amount,
			Type:                transaction.Type,
			CategoryID:          transaction.CategoryID,
			Notes:               transaction.Notes,
			IsRecurring:         transaction.IsRecurring,
			CreatedAt:           transaction.CreatedAt,
			UpdatedAt:           transaction.UpdatedAt,
			BillingCycle:        transaction.BillingCycle,
			IsExpandedBill:      transaction.ExpandedAt != nil,
			InstallmentCurrent:  transaction.InstallmentCurrent,
			InstallmentTotal:    transaction.InstallmentTotal,
			CreditCardPaymentID: transaction.CreditCardPaymentID,
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

// autoCategorize attempts to match the transaction description against the user's category rules.
// It returns the matched category ID and category entity if a match is found, or nil if no match.
// Rules are already sorted by priority (highest first) from the repository.
// This method does not fail the transaction creation if rule matching fails - it logs and continues.
func (uc *CreateTransactionUseCase) autoCategorize(
	ctx context.Context,
	userID uuid.UUID,
	description string,
) (*uuid.UUID, *entity.Category) {
	// Fetch all active rules for the user, sorted by priority (descending)
	rules, err := uc.categoryRuleRepo.FindActiveByOwner(ctx, entity.OwnerTypeUser, userID)
	if err != nil {
		slog.Debug("Failed to fetch category rules for auto-categorization",
			"userID", userID,
			"error", err,
		)
		return nil, nil
	}

	if len(rules) == 0 {
		return nil, nil
	}

	// Try to match each rule against the description
	for _, rule := range rules {
		// Create case-insensitive regex pattern
		pattern := "(?i)" + rule.Pattern
		re, err := regexp.Compile(pattern)
		if err != nil {
			slog.Debug("Invalid regex pattern in category rule",
				"ruleID", rule.ID,
				"pattern", rule.Pattern,
				"error", err,
			)
			continue
		}

		if re.MatchString(description) {
			// Found a match - fetch the category to return with the transaction
			category, err := uc.categoryRepo.FindByID(ctx, rule.CategoryID)
			if err != nil {
				slog.Debug("Failed to fetch category for matched rule",
					"ruleID", rule.ID,
					"categoryID", rule.CategoryID,
					"error", err,
				)
				continue
			}

			slog.Debug("Auto-categorized transaction",
				"userID", userID,
				"description", description,
				"ruleID", rule.ID,
				"categoryID", rule.CategoryID,
				"categoryName", category.Name,
			)

			return &rule.CategoryID, category
		}
	}

	return nil, nil
}

// Package creditcard contains credit card import-related use cases.
package creditcard

import (
	"context"
	"log/slog"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// ImportTransactionsInput represents the input for importing CC transactions.
type ImportTransactionsInput struct {
	UserID            uuid.UUID
	BillingCycle      string
	BillPaymentID     uuid.UUID
	Transactions      []CCTransactionInput
	ApplyAutoCategory bool
}

// ImportedTransactionSummary represents a summary of an imported transaction.
type ImportedTransactionSummary struct {
	ID          uuid.UUID
	Date        time.Time
	Description string
	Amount      decimal.Decimal
	CategoryID  *uuid.UUID
}

// ImportTransactionsOutput represents the output of CC import operation.
type ImportTransactionsOutput struct {
	ImportedCount      int
	CategorizedCount   int
	BillPaymentID      uuid.UUID
	BillingCycle       string
	OriginalBillAmount decimal.Decimal
	ImportedAt         time.Time
	Transactions       []ImportedTransactionSummary
}

// ImportTransactionsUseCase handles the CC import logic.
type ImportTransactionsUseCase struct {
	transactionRepo  adapter.TransactionRepository
	categoryRepo     adapter.CategoryRepository
	categoryRuleRepo adapter.CategoryRuleRepository
}

// NewImportTransactionsUseCase creates a new ImportTransactionsUseCase instance.
func NewImportTransactionsUseCase(
	transactionRepo adapter.TransactionRepository,
	categoryRepo adapter.CategoryRepository,
	categoryRuleRepo adapter.CategoryRuleRepository,
) *ImportTransactionsUseCase {
	return &ImportTransactionsUseCase{
		transactionRepo:  transactionRepo,
		categoryRepo:     categoryRepo,
		categoryRuleRepo: categoryRuleRepo,
	}
}

// Execute performs the CC import operation.
func (uc *ImportTransactionsUseCase) Execute(ctx context.Context, input ImportTransactionsInput) (*ImportTransactionsOutput, error) {
	// Validate billing cycle format
	if !billingCycleRegex.MatchString(input.BillingCycle) {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeInvalidBillingCycle,
			"billing cycle must be in YYYY-MM format",
			domainerror.ErrInvalidBillingCycle,
		)
	}

	// Validate transactions
	if len(input.Transactions) == 0 {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeEmptyCCTransactions,
			"at least one transaction is required",
			domainerror.ErrEmptyCCTransactions,
		)
	}

	// Verify bill payment exists and belongs to user
	billPayment, err := uc.transactionRepo.FindBillPaymentByID(ctx, input.BillPaymentID, input.UserID)
	if err != nil {
		if err == domainerror.ErrBillPaymentNotFound {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeBillPaymentNotFound,
				"bill payment transaction not found",
				domainerror.ErrBillPaymentNotFound,
			)
		}
		return nil, err
	}

	// Check if bill is already expanded
	isExpanded, err := uc.transactionRepo.IsBillExpanded(ctx, input.BillPaymentID)
	if err != nil {
		return nil, err
	}
	if isExpanded {
		return nil, domainerror.NewTransactionError(
			domainerror.ErrCodeBillAlreadyExpanded,
			"bill is already expanded with credit card transactions",
			domainerror.ErrBillAlreadyExpanded,
		)
	}

	// Prepare category rules for auto-categorization if enabled
	var categoryRules []*entity.CategoryRule
	if input.ApplyAutoCategory {
		rules, err := uc.categoryRuleRepo.FindActiveByOwner(ctx, entity.OwnerTypeUser, input.UserID)
		if err != nil {
			slog.Warn("Failed to fetch category rules for auto-categorization",
				"userID", input.UserID,
				"error", err,
			)
		} else {
			categoryRules = rules
		}
	}

	// Create CC transaction entities
	now := time.Now().UTC()
	var transactions []*entity.Transaction
	var transactionSummaries []ImportedTransactionSummary
	categorizedCount := 0
	paymentReceivedRegex := regexp.MustCompile(PaymentReceivedPattern)

	for _, txnInput := range input.Transactions {
		// Determine if this is a "Pagamento recebido" entry
		isPaymentReceived := paymentReceivedRegex.MatchString(txnInput.Description)

		// Create transaction entity
		txn := &entity.Transaction{
			ID:                  uuid.New(),
			UserID:              input.UserID,
			Date:                txnInput.Date,
			Description:         txnInput.Description,
			Amount:              txnInput.Amount,
			Type:                entity.TransactionTypeExpense, // CC transactions are expenses
			CreditCardPaymentID: &input.BillPaymentID,
			BillingCycle:        input.BillingCycle,
			InstallmentCurrent:  txnInput.InstallmentCurrent,
			InstallmentTotal:    txnInput.InstallmentTotal,
			IsHidden:            isPaymentReceived, // Hide "Pagamento recebido" entries
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		// Apply auto-categorization if enabled and not a payment received entry
		if input.ApplyAutoCategory && !isPaymentReceived {
			categoryID, _ := uc.autoCategorize(ctx, txnInput.Description, categoryRules)
			if categoryID != nil {
				txn.CategoryID = categoryID
				categorizedCount++
			}
		}

		transactions = append(transactions, txn)
		transactionSummaries = append(transactionSummaries, ImportedTransactionSummary{
			ID:          txn.ID,
			Date:        txn.Date,
			Description: txn.Description,
			Amount:      txn.Amount,
			CategoryID:  txn.CategoryID,
		})
	}

	// Save the original bill amount before zeroing
	originalBillAmount := billPayment.Amount.Abs()

	// Create all CC transactions and update bill payment in a single transaction
	if err := uc.transactionRepo.BulkCreateCCTransactions(
		ctx,
		transactions,
		input.BillPaymentID,
		originalBillAmount,
	); err != nil {
		return nil, err
	}

	return &ImportTransactionsOutput{
		ImportedCount:      len(transactions),
		CategorizedCount:   categorizedCount,
		BillPaymentID:      input.BillPaymentID,
		BillingCycle:       input.BillingCycle,
		OriginalBillAmount: originalBillAmount,
		ImportedAt:         now,
		Transactions:       transactionSummaries,
	}, nil
}

// autoCategorize attempts to match the description against category rules.
func (uc *ImportTransactionsUseCase) autoCategorize(
	ctx context.Context,
	description string,
	rules []*entity.CategoryRule,
) (*uuid.UUID, *entity.Category) {
	if len(rules) == 0 {
		return nil, nil
	}

	for _, rule := range rules {
		// Create case-insensitive regex pattern
		pattern := "(?i)" + rule.Pattern
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		if re.MatchString(description) {
			// Found a match
			category, err := uc.categoryRepo.FindByID(ctx, rule.CategoryID)
			if err != nil {
				continue
			}

			slog.Debug("Auto-categorized CC transaction",
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

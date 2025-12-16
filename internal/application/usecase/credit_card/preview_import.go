// Package creditcard contains credit card import-related use cases.
package creditcard

import (
	"context"
	"errors"
	"math"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

const (
	// BillingCyclePattern is the regex pattern for valid billing cycle format (YYYY-MM).
	BillingCyclePattern = `^\d{4}-(0[1-9]|1[0-2])$`
	// AmountTolerancePercent is the percentage tolerance for amount matching (1%).
	AmountTolerancePercent = 0.01
	// AmountToleranceAbsolute is the absolute tolerance for amount matching (R$ 10.00).
	AmountToleranceAbsolute = 10.0
	// DateProximityDays is the maximum days difference for date matching.
	// Set to 10 to accommodate CC statement dates vs bank payment dates (typically 5-7 days apart).
	DateProximityDays = 10
	// PaymentReceivedPattern matches "Pagamento recebido" entries in CC statements.
	PaymentReceivedPattern = `(?i)pagamento\s+recebido`
)

var billingCycleRegex = regexp.MustCompile(BillingCyclePattern)

// CCTransactionInput represents a parsed credit card transaction line.
type CCTransactionInput struct {
	Date               time.Time
	Description        string
	Amount             decimal.Decimal
	InstallmentCurrent *int
	InstallmentTotal   *int
}

// BillMatch represents a potential match between CC payment and bank bill payment.
type BillMatch struct {
	BillPaymentID     uuid.UUID
	BillPaymentDate   time.Time
	BillPaymentAmount decimal.Decimal
	BillDescription   string
	CCPaymentDate     time.Time
	CCPaymentAmount   decimal.Decimal
	MatchScore        float64
}

// PreviewImportInput represents the input for previewing CC import.
type PreviewImportInput struct {
	UserID       uuid.UUID
	BillingCycle string
	Transactions []CCTransactionInput
}

// PreviewImportOutput represents the output of CC import preview.
type PreviewImportOutput struct {
	BillingCycle          string
	TotalTransactions     int
	TotalAmount           decimal.Decimal
	PotentialMatches      []BillMatch
	TransactionsToImport  []CCTransactionInput
	PaymentReceivedAmount decimal.Decimal
	HasExistingImport     bool
}

// PreviewImportUseCase handles the CC import preview logic.
type PreviewImportUseCase struct {
	transactionRepo adapter.TransactionRepository
}

// NewPreviewImportUseCase creates a new PreviewImportUseCase instance.
func NewPreviewImportUseCase(transactionRepo adapter.TransactionRepository) *PreviewImportUseCase {
	return &PreviewImportUseCase{
		transactionRepo: transactionRepo,
	}
}

// Execute performs the CC import preview.
func (uc *PreviewImportUseCase) Execute(ctx context.Context, input PreviewImportInput) (*PreviewImportOutput, error) {
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

	// Check if billing cycle already has imported transactions
	status, err := uc.transactionRepo.GetCreditCardStatus(ctx, input.UserID, input.BillingCycle)
	if err != nil {
		// Wrap database errors in TransactionError for proper error handling
		var txnErr *domainerror.TransactionError
		if !errors.As(err, &txnErr) {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeInternalError,
				"failed to check credit card status",
				err,
			)
		}
		return nil, err
	}

	// Calculate totals and separate "Pagamento recebido" entries
	var totalAmount decimal.Decimal
	var paymentReceivedAmount decimal.Decimal
	var transactionsToImport []CCTransactionInput
	var ccPaymentDate time.Time
	var ccPaymentAmount decimal.Decimal

	paymentReceivedRegex := regexp.MustCompile(PaymentReceivedPattern)

	for _, txn := range input.Transactions {
		if paymentReceivedRegex.MatchString(txn.Description) {
			// This is the "Pagamento recebido" entry - typically negative in CC statement
			// Use Abs() here since we want the positive reference amount for matching
			paymentReceivedAmount = paymentReceivedAmount.Add(txn.Amount.Abs())
			ccPaymentDate = txn.Date
			ccPaymentAmount = txn.Amount
		} else {
			// Regular CC transaction - preserve sign for algebraic sum
			// Refunds (negative amounts like "Estorno de compra") should subtract from total
			totalAmount = totalAmount.Add(txn.Amount)
			transactionsToImport = append(transactionsToImport, txn)
		}
	}

	// Find potential bill payment matches
	// Search in a date range around the billing cycle
	startDate, endDate := calculateSearchDateRange(input.BillingCycle)
	potentialBills, err := uc.transactionRepo.FindPotentialBillPayments(ctx, input.UserID, startDate, endDate)
	if err != nil {
		// Wrap database errors in TransactionError for proper error handling
		var txnErr *domainerror.TransactionError
		if !errors.As(err, &txnErr) {
			return nil, domainerror.NewTransactionError(
				domainerror.ErrCodeInternalError,
				"failed to find potential bill payments",
				err,
			)
		}
		return nil, err
	}

	// Match bills with CC payment amount
	matches := uc.matchBillPayments(potentialBills, ccPaymentDate, ccPaymentAmount)

	return &PreviewImportOutput{
		BillingCycle:          input.BillingCycle,
		TotalTransactions:     len(transactionsToImport),
		TotalAmount:           totalAmount,
		PotentialMatches:      matches,
		TransactionsToImport:  transactionsToImport,
		PaymentReceivedAmount: paymentReceivedAmount,
		HasExistingImport:     status.IsExpanded,
	}, nil
}

// matchBillPayments finds and scores potential bill payment matches.
func (uc *PreviewImportUseCase) matchBillPayments(
	bills []*entity.Transaction,
	ccPaymentDate time.Time,
	ccPaymentAmount decimal.Decimal,
) []BillMatch {
	var matches []BillMatch
	ccAmountAbs := ccPaymentAmount.Abs()

	for _, bill := range bills {
		billAmountAbs := bill.Amount.Abs()

		// Calculate amount difference
		amountDiff := billAmountAbs.Sub(ccAmountAbs).Abs()
		percentDiff := float64(0)
		if !billAmountAbs.IsZero() {
			percentDiff, _ = amountDiff.Div(billAmountAbs).Float64()
		}
		absoluteDiff, _ := amountDiff.Float64()

		// Calculate date difference
		daysDiff := int(math.Abs(bill.Date.Sub(ccPaymentDate).Hours() / 24))

		// Check if within tolerance
		isAmountMatch := percentDiff <= AmountTolerancePercent || absoluteDiff <= AmountToleranceAbsolute
		isDateMatch := daysDiff <= DateProximityDays

		if isAmountMatch && isDateMatch {
			// Calculate match score (higher is better)
			// Prefer closer amounts and dates
			amountScore := 1.0 - math.Min(percentDiff/AmountTolerancePercent, 1.0)
			dateScore := 1.0 - float64(daysDiff)/float64(DateProximityDays)
			matchScore := (amountScore * 0.7) + (dateScore * 0.3) // Weight amount more heavily

			matches = append(matches, BillMatch{
				BillPaymentID:     bill.ID,
				BillPaymentDate:   bill.Date,
				BillPaymentAmount: bill.Amount,
				BillDescription:   bill.Description,
				CCPaymentDate:     ccPaymentDate,
				CCPaymentAmount:   ccPaymentAmount,
				MatchScore:        matchScore,
			})
		}
	}

	// Sort by match score (descending)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].MatchScore > matches[j].MatchScore
	})

	return matches
}

// calculateSearchDateRange calculates the date range for searching bill payments.
// The range is typically the billing cycle month plus/minus a buffer.
func calculateSearchDateRange(billingCycle string) (time.Time, time.Time) {
	// Parse billing cycle (YYYY-MM)
	year, month := parseBillingCycle(billingCycle)

	// Start from the first day of the previous month
	startDate := time.Date(year, time.Month(month-1), 1, 0, 0, 0, 0, time.UTC)
	if month == 1 {
		startDate = time.Date(year-1, 12, 1, 0, 0, 0, 0, time.UTC)
	}

	// End at the last day of the next month
	endMonth := time.Month(month + 1)
	endYear := year
	if month == 12 {
		endMonth = 1
		endYear = year + 1
	}
	endDate := time.Date(endYear, endMonth+1, 0, 23, 59, 59, 0, time.UTC)

	return startDate, endDate
}

// parseBillingCycle parses a billing cycle string (YYYY-MM) into year and month.
func parseBillingCycle(billingCycle string) (int, int) {
	// Format is already validated by regex
	t, _ := time.Parse("2006-01", billingCycle)
	return t.Year(), int(t.Month())
}


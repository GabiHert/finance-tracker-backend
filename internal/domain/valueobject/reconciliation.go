// Package valueobject contains domain value objects for the Finance Tracker system.
package valueobject

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Confidence represents the confidence level of a bill match.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// PotentialMatch represents a potential bill payment match for CC transactions.
type PotentialMatch struct {
	BillID                  uuid.UUID
	BillDate                time.Time
	BillDescription         string
	BillAmount              decimal.Decimal
	CategoryName            *string
	Confidence              Confidence
	AmountDifference        decimal.Decimal // CC total - Bill amount (negative if bill > CC)
	AmountDifferencePercent decimal.Decimal
	Score                   float64 // For ranking multiple matches (0-1.0)
}

// PendingCycle represents a billing cycle with pending (unlinked) CC transactions.
type PendingCycle struct {
	BillingCycle     string
	DisplayName      string // e.g., "Nov/2024"
	TransactionCount int
	TotalAmount      decimal.Decimal
	OldestDate       time.Time
	NewestDate       time.Time
	PotentialBills   []PotentialMatch
}

// LinkedCycle represents a billing cycle with linked CC transactions.
type LinkedCycle struct {
	BillingCycle     string
	DisplayName      string
	TransactionCount int
	TotalAmount      decimal.Decimal
	Bill             LinkedBill
	AmountDifference decimal.Decimal
	HasMismatch      bool
}

// LinkedBill contains information about the linked bill payment.
type LinkedBill struct {
	ID             uuid.UUID
	Date           time.Time
	Description    string
	OriginalAmount decimal.Decimal
	CategoryName   *string
}

// ReconciliationSummary contains summary statistics for reconciliation.
type ReconciliationSummary struct {
	TotalPending int
	TotalLinked  int
	MonthsCovered int
}

// ReconciliationResult represents the result of a reconciliation operation.
type ReconciliationResult struct {
	AutoLinked        []AutoLinkedCycle
	RequiresSelection []PendingWithMatches
	NoMatch           []NoMatchCycle
	Summary           ReconciliationResultSummary
}

// AutoLinkedCycle represents a billing cycle that was automatically linked.
type AutoLinkedCycle struct {
	BillingCycle     string
	BillID           uuid.UUID
	BillDescription  string
	TransactionCount int
	Confidence       Confidence
	AmountDifference decimal.Decimal
}

// PendingWithMatches represents a pending cycle with multiple potential matches.
type PendingWithMatches struct {
	BillingCycle   string
	PotentialBills []PotentialMatch
}

// NoMatchCycle represents a billing cycle with no matching bills.
type NoMatchCycle struct {
	BillingCycle     string
	TransactionCount int
	TotalAmount      decimal.Decimal
}

// ReconciliationResultSummary contains counts from a reconciliation operation.
type ReconciliationResultSummary struct {
	AutoLinked        int
	RequiresSelection int
	NoMatch           int
}

// LinkResult represents the result of manually linking CC to a bill.
type LinkResult struct {
	BillingCycle       string
	BillID             uuid.UUID
	TransactionsLinked int
	AmountDifference   decimal.Decimal
	HasMismatch        bool
}

// AutoReconciliationResult represents the result of auto-reconciliation on bill creation.
type AutoReconciliationResult struct {
	Triggered          bool
	LinkedCycle        *string
	TransactionsLinked int
	Confidence         Confidence
}

// CalculateConfidence determines the confidence level for a match based on amount difference.
func CalculateConfidence(config MatchingConfig, ccTotal, billAmount decimal.Decimal) Confidence {
	diff := ccTotal.Sub(billAmount).Abs()

	// Check high confidence thresholds (either absolute or percentage)
	highAbsolute := decimal.NewFromInt(config.HighConfidenceAbsolute)
	if diff.LessThanOrEqual(highAbsolute) {
		return ConfidenceHigh
	}

	if !billAmount.IsZero() {
		percentDiff := diff.Div(billAmount.Abs())
		if percentDiff.LessThanOrEqual(config.HighConfidencePercent) {
			return ConfidenceHigh
		}
	}

	// Check medium confidence thresholds
	medAbsolute := decimal.NewFromInt(config.MedConfidenceAbsolute)
	if diff.LessThanOrEqual(medAbsolute) {
		return ConfidenceMedium
	}

	if !billAmount.IsZero() {
		percentDiff := diff.Div(billAmount.Abs())
		if percentDiff.LessThanOrEqual(config.MedConfidencePercent) {
			return ConfidenceMedium
		}
	}

	return ConfidenceLow
}

// FormatBillingCycleDisplay formats a billing cycle (YYYY-MM) for display (e.g., "Nov/2024").
func FormatBillingCycleDisplay(billingCycle string) string {
	if len(billingCycle) != 7 {
		return billingCycle
	}

	monthNames := []string{
		"Jan", "Fev", "Mar", "Abr", "Mai", "Jun",
		"Jul", "Ago", "Set", "Out", "Nov", "Dez",
	}

	year := billingCycle[:4]
	monthStr := billingCycle[5:7]

	// Parse month
	month := 0
	if monthStr[0] == '0' {
		month = int(monthStr[1] - '0')
	} else {
		month = int(monthStr[0]-'0')*10 + int(monthStr[1]-'0')
	}

	if month < 1 || month > 12 {
		return billingCycle
	}

	return monthNames[month-1] + "/" + year
}

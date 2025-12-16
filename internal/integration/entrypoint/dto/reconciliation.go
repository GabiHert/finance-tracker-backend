// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/domain/valueobject"
)

// PotentialBillDTO represents a potential bill match for reconciliation.
type PotentialBillDTO struct {
	BillID                  string `json:"bill_id"`
	BillDate                string `json:"bill_date"`
	BillDescription         string `json:"bill_description"`
	BillAmount              string `json:"bill_amount"`
	CategoryName            string `json:"category_name,omitempty"`
	Confidence              string `json:"confidence"`
	AmountDifference        string `json:"amount_difference"`
	AmountDifferencePercent string `json:"amount_difference_percent"`
	Score                   float64 `json:"score"`
}

// PendingCycleDTO represents a pending billing cycle with potential matches.
type PendingCycleDTO struct {
	BillingCycle     string             `json:"billing_cycle"`
	DisplayName      string             `json:"display_name"`
	TransactionCount int                `json:"transaction_count"`
	TotalAmount      string             `json:"total_amount"`
	OldestDate       string             `json:"oldest_date"`
	NewestDate       string             `json:"newest_date"`
	PotentialBills   []PotentialBillDTO `json:"potential_bills"`
}

// LinkedBillDTO contains information about a linked bill payment.
type LinkedBillDTO struct {
	ID             string `json:"id"`
	Date           string `json:"date"`
	Description    string `json:"description"`
	OriginalAmount string `json:"original_amount"`
	CategoryName   string `json:"category_name,omitempty"`
}

// LinkedCycleDTO represents a linked billing cycle.
type LinkedCycleDTO struct {
	BillingCycle     string        `json:"billing_cycle"`
	DisplayName      string        `json:"display_name"`
	TransactionCount int           `json:"transaction_count"`
	TotalAmount      string        `json:"total_amount"`
	Bill             LinkedBillDTO `json:"bill"`
	AmountDifference string        `json:"amount_difference"`
	HasMismatch      bool          `json:"has_mismatch"`
}

// ReconciliationSummaryDTO contains summary statistics.
type ReconciliationSummaryDTO struct {
	TotalPending  int `json:"total_pending"`
	TotalLinked   int `json:"total_linked"`
	MonthsCovered int `json:"months_covered"`
}

// GetPendingResponseDTO represents the response for GET /reconciliation/pending.
type GetPendingResponseDTO struct {
	PendingCycles []PendingCycleDTO        `json:"pending_cycles"`
	Summary       ReconciliationSummaryDTO `json:"summary"`
}

// GetLinkedResponseDTO represents the response for GET /reconciliation/linked.
type GetLinkedResponseDTO struct {
	LinkedCycles []LinkedCycleDTO         `json:"linked_cycles"`
	Summary      ReconciliationSummaryDTO `json:"summary"`
}

// GetSummaryResponseDTO represents the response for GET /reconciliation/summary.
type GetSummaryResponseDTO struct {
	TotalPending  int `json:"total_pending"`
	TotalLinked   int `json:"total_linked"`
	MonthsCovered int `json:"months_covered"`
}

// ManualLinkRequestDTO represents the request for POST /reconciliation/link.
type ManualLinkRequestDTO struct {
	BillingCycle  string `json:"billing_cycle" binding:"required"`
	BillPaymentID string `json:"bill_payment_id" binding:"required"`
	Force         bool   `json:"force"`
}

// ManualLinkResponseDTO represents the response for POST /reconciliation/link.
type ManualLinkResponseDTO struct {
	BillingCycle       string `json:"billing_cycle"`
	BillPaymentID      string `json:"bill_payment_id"`
	TransactionsLinked int    `json:"transactions_linked"`
	AmountDifference   string `json:"amount_difference"`
	HasMismatch        bool   `json:"has_mismatch"`
}

// UnlinkRequestDTO represents the request for POST /reconciliation/unlink.
type UnlinkRequestDTO struct {
	BillingCycle string `json:"billing_cycle" binding:"required"`
}

// UnlinkResponseDTO represents the response for POST /reconciliation/unlink.
type UnlinkResponseDTO struct {
	BillingCycle string `json:"billing_cycle"`
	Success      bool   `json:"success"`
}

// TriggerReconciliationRequestDTO represents the request for POST /reconciliation/trigger.
type TriggerReconciliationRequestDTO struct {
	BillingCycle string `json:"billing_cycle"` // Optional - if empty, reconcile all pending
}

// AutoLinkedCycleDTO represents an auto-linked cycle in the response.
type AutoLinkedCycleDTO struct {
	BillingCycle     string `json:"billing_cycle"`
	BillID           string `json:"bill_id"`
	BillDescription  string `json:"bill_description"`
	TransactionCount int    `json:"transaction_count"`
	Confidence       string `json:"confidence"`
	AmountDifference string `json:"amount_difference"`
}

// PendingWithMatchesDTO represents a pending cycle with multiple matches.
type PendingWithMatchesDTO struct {
	BillingCycle   string             `json:"billing_cycle"`
	PotentialBills []PotentialBillDTO `json:"potential_bills"`
}

// NoMatchCycleDTO represents a cycle with no matching bills.
type NoMatchCycleDTO struct {
	BillingCycle     string `json:"billing_cycle"`
	TransactionCount int    `json:"transaction_count"`
	TotalAmount      string `json:"total_amount"`
}

// ReconciliationResultSummaryDTO contains counts from reconciliation.
type ReconciliationResultSummaryDTO struct {
	AutoLinked        int `json:"auto_linked"`
	RequiresSelection int `json:"requires_selection"`
	NoMatch           int `json:"no_match"`
}

// TriggerReconciliationResponseDTO represents the response for POST /reconciliation/trigger.
type TriggerReconciliationResponseDTO struct {
	AutoLinked        []AutoLinkedCycleDTO           `json:"auto_linked"`
	RequiresSelection []PendingWithMatchesDTO        `json:"requires_selection"`
	NoMatch           []NoMatchCycleDTO              `json:"no_match"`
	Summary           ReconciliationResultSummaryDTO `json:"summary"`
}

// ToPotentialBillDTO converts domain data to DTO.
func ToPotentialBillDTO(
	billID uuid.UUID,
	billDate time.Time,
	billDescription string,
	billAmount decimal.Decimal,
	categoryName *string,
	confidence valueobject.Confidence,
	amountDiff decimal.Decimal,
	amountDiffPercent decimal.Decimal,
	score float64,
) PotentialBillDTO {
	dto := PotentialBillDTO{
		BillID:                  billID.String(),
		BillDate:                billDate.Format("2006-01-02"),
		BillDescription:         billDescription,
		BillAmount:              billAmount.String(),
		Confidence:              string(confidence),
		AmountDifference:        amountDiff.String(),
		AmountDifferencePercent: amountDiffPercent.StringFixed(4),
		Score:                   score,
	}
	if categoryName != nil {
		dto.CategoryName = *categoryName
	}
	return dto
}

// ToPendingCycleDTO converts domain data to DTO.
func ToPendingCycleDTO(
	billingCycle string,
	displayName string,
	transactionCount int,
	totalAmount decimal.Decimal,
	oldestDate time.Time,
	newestDate time.Time,
	potentialBills []PotentialBillDTO,
) PendingCycleDTO {
	return PendingCycleDTO{
		BillingCycle:     billingCycle,
		DisplayName:      displayName,
		TransactionCount: transactionCount,
		TotalAmount:      totalAmount.String(),
		OldestDate:       oldestDate.Format("2006-01-02"),
		NewestDate:       newestDate.Format("2006-01-02"),
		PotentialBills:   potentialBills,
	}
}

// ToLinkedCycleDTO converts domain data to DTO.
func ToLinkedCycleDTO(
	billingCycle string,
	displayName string,
	transactionCount int,
	totalAmount decimal.Decimal,
	billID uuid.UUID,
	billDate time.Time,
	billDescription string,
	billOriginalAmount decimal.Decimal,
	categoryName *string,
	amountDiff decimal.Decimal,
	hasMismatch bool,
) LinkedCycleDTO {
	billDTO := LinkedBillDTO{
		ID:             billID.String(),
		Date:           billDate.Format("2006-01-02"),
		Description:    billDescription,
		OriginalAmount: billOriginalAmount.String(),
	}
	if categoryName != nil {
		billDTO.CategoryName = *categoryName
	}

	return LinkedCycleDTO{
		BillingCycle:     billingCycle,
		DisplayName:      displayName,
		TransactionCount: transactionCount,
		TotalAmount:      totalAmount.String(),
		Bill:             billDTO,
		AmountDifference: amountDiff.String(),
		HasMismatch:      hasMismatch,
	}
}

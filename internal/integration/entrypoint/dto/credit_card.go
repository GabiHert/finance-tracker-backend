// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreditCardTransactionDTO represents a parsed credit card transaction line.
type CreditCardTransactionDTO struct {
	Date               string  `json:"date"`
	Description        string  `json:"description"`
	Amount             float64 `json:"amount"`
	InstallmentCurrent *int    `json:"installment_current,omitempty"`
	InstallmentTotal   *int    `json:"installment_total,omitempty"`
}

// BillMatchDTO represents a potential match between CC payment and bank bill payment.
type BillMatchDTO struct {
	BillPaymentID     string  `json:"bill_payment_id"`
	BillPaymentDate   string  `json:"bill_payment_date"`
	BillPaymentAmount string  `json:"bill_payment_amount"`
	BillDescription   string  `json:"bill_description"`
	CCPaymentDate     string  `json:"cc_payment_date"`
	CCPaymentAmount   string  `json:"cc_payment_amount"`
	AmountDifference  string  `json:"amount_difference"`
	DaysDifference    int     `json:"days_difference"`
	MatchScore        float64 `json:"match_score"`
}

// ImportPreviewRequestDTO represents the request for previewing CC import.
type ImportPreviewRequestDTO struct {
	BillingCycle string                     `json:"billing_cycle" binding:"required"` // Format: "YYYY-MM"
	Transactions []CreditCardTransactionDTO `json:"transactions" binding:"required,min=1"`
}

// ImportPreviewResponseDTO represents the response for CC import preview.
type ImportPreviewResponseDTO struct {
	BillingCycle          string                     `json:"billing_cycle"`
	TotalTransactions     int                        `json:"total_transactions"`
	TotalAmount           string                     `json:"total_amount"`
	PotentialMatches      []BillMatchDTO             `json:"potential_matches"`
	TransactionsToImport  []CreditCardTransactionDTO `json:"transactions_to_import"`
	PaymentReceivedAmount string                     `json:"payment_received_amount"` // "Pagamento recebido" total
	HasExistingImport     bool                       `json:"has_existing_import"`     // If billing cycle already imported
}

// ImportRequestDTO represents the request for importing CC transactions.
type ImportRequestDTO struct {
	BillingCycle      string                     `json:"billing_cycle" binding:"required"` // Format: "YYYY-MM"
	BillPaymentID     string                     `json:"bill_payment_id"`                  // Optional - if empty, imports as standalone CC transactions
	Transactions      []CreditCardTransactionDTO `json:"transactions" binding:"required,min=1"`
	ApplyAutoCategory bool                       `json:"apply_auto_category"` // Whether to apply category rules
}

// ImportResultDTO represents the result of CC import operation.
type ImportResultDTO struct {
	ImportedCount      int                          `json:"imported_count"`
	CategorizedCount   int                          `json:"categorized_count"`
	BillPaymentID      string                       `json:"bill_payment_id"`
	BillingCycle       string                       `json:"billing_cycle"`
	OriginalBillAmount string                       `json:"original_bill_amount"`
	ImportedAt         time.Time                    `json:"imported_at"`
	Transactions       []ImportedTransactionSummary `json:"transactions"`
}

// ImportedTransactionSummary represents a summary of an imported transaction.
type ImportedTransactionSummary struct {
	ID          string  `json:"id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      string  `json:"amount"`
	CategoryID  *string `json:"category_id,omitempty"`
}

// CollapseRequestDTO represents the request for collapsing CC expansion.
type CollapseRequestDTO struct {
	BillPaymentID string `json:"bill_payment_id" binding:"required"`
}

// CollapseResultDTO represents the result of collapse operation.
type CollapseResultDTO struct {
	BillPaymentID       string `json:"bill_payment_id"`
	RestoredAmount      string `json:"restored_amount"`
	DeletedTransactions int    `json:"deleted_transactions"`
	CollapsedAt         time.Time `json:"collapsed_at"`
}

// CreditCardStatusRequestDTO represents the query parameters for CC status.
type CreditCardStatusRequestDTO struct {
	BillingCycle string `form:"billing_cycle"` // Optional: "YYYY-MM", defaults to current month
}

// CreditCardStatusDTO represents the credit card status for a billing cycle.
type CreditCardStatusDTO struct {
	BillingCycle        string                    `json:"billing_cycle"`
	IsExpanded          bool                      `json:"is_expanded"`
	BillPaymentID       *string                   `json:"bill_payment_id,omitempty"`
	BillPaymentDate     *string                   `json:"bill_payment_date,omitempty"`
	OriginalAmount      *string                   `json:"original_amount,omitempty"`
	CurrentAmount       *string                   `json:"current_amount,omitempty"`
	LinkedTransactions  int                       `json:"linked_transactions"`
	ExpandedAt          *time.Time                `json:"expanded_at,omitempty"`
	TransactionsSummary []CCTransactionSummaryDTO `json:"transactions_summary,omitempty"`
}

// CCTransactionSummaryDTO represents a summary of a CC transaction.
type CCTransactionSummaryDTO struct {
	ID          string  `json:"id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      string  `json:"amount"`
	CategoryID  *string `json:"category_id,omitempty"`
	IsHidden    bool    `json:"is_hidden"`
}

// ToCreditCardTransactionDTO converts use case output to DTO.
func ToCreditCardTransactionDTO(
	id uuid.UUID,
	date time.Time,
	description string,
	amount decimal.Decimal,
	categoryID *uuid.UUID,
	isHidden bool,
) CCTransactionSummaryDTO {
	dto := CCTransactionSummaryDTO{
		ID:          id.String(),
		Date:        date.Format("2006-01-02"),
		Description: description,
		Amount:      amount.String(),
		IsHidden:    isHidden,
	}
	if categoryID != nil {
		catIDStr := categoryID.String()
		dto.CategoryID = &catIDStr
	}
	return dto
}

// ToBillMatchDTO converts use case output to DTO.
func ToBillMatchDTO(
	billPaymentID uuid.UUID,
	billPaymentDate time.Time,
	billPaymentAmount decimal.Decimal,
	billDescription string,
	ccPaymentDate time.Time,
	ccPaymentAmount decimal.Decimal,
	matchScore float64,
) BillMatchDTO {
	amountDiff := billPaymentAmount.Sub(ccPaymentAmount.Abs())
	daysDiff := int(billPaymentDate.Sub(ccPaymentDate).Hours() / 24)
	if daysDiff < 0 {
		daysDiff = -daysDiff
	}

	return BillMatchDTO{
		BillPaymentID:     billPaymentID.String(),
		BillPaymentDate:   billPaymentDate.Format("2006-01-02"),
		BillPaymentAmount: billPaymentAmount.String(),
		BillDescription:   billDescription,
		CCPaymentDate:     ccPaymentDate.Format("2006-01-02"),
		CCPaymentAmount:   ccPaymentAmount.Abs().String(),
		AmountDifference:  amountDiff.Abs().String(),
		DaysDifference:    daysDiff,
		MatchScore:        matchScore,
	}
}

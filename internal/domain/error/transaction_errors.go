// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// Transaction domain errors.
var (
	// ErrTransactionNotFound is returned when a transaction is not found in the system.
	ErrTransactionNotFound = errors.New("transaction not found")

	// ErrNotAuthorizedToModifyTransaction is returned when user is not authorized to modify a transaction.
	ErrNotAuthorizedToModifyTransaction = errors.New("not authorized to modify transaction")

	// ErrInvalidTransactionType is returned when the transaction type is invalid.
	ErrInvalidTransactionType = errors.New("invalid transaction type")

	// ErrInvalidTransactionDate is returned when the transaction date is invalid.
	ErrInvalidTransactionDate = errors.New("invalid transaction date")

	// ErrInvalidTransactionAmount is returned when the transaction amount is invalid.
	ErrInvalidTransactionAmount = errors.New("invalid transaction amount")

	// ErrCategoryNotFoundForTransaction is returned when the specified category is not found.
	ErrCategoryNotFoundForTransaction = errors.New("category not found")

	// ErrCategoryNotOwnedByUser is returned when the category does not belong to the user.
	ErrCategoryNotOwnedByUser = errors.New("category does not belong to user")

	// ErrDescriptionTooLong is returned when the transaction description exceeds the maximum length.
	ErrDescriptionTooLong = errors.New("description too long")

	// ErrNotesTooLong is returned when the transaction notes exceed the maximum length.
	ErrNotesTooLong = errors.New("notes too long")

	// ErrEmptyTransactionIDs is returned when an empty list of transaction IDs is provided.
	ErrEmptyTransactionIDs = errors.New("transaction IDs list cannot be empty")

	// ErrTransactionIDsNotFound is returned when one or more transaction IDs are not found.
	ErrTransactionIDsNotFound = errors.New("one or more transactions not found")

	// Credit card import errors.

	// ErrInvalidBillingCycle is returned when the billing cycle format is invalid.
	ErrInvalidBillingCycle = errors.New("invalid billing cycle format")

	// ErrBillPaymentNotFound is returned when the bill payment transaction is not found.
	ErrBillPaymentNotFound = errors.New("bill payment transaction not found")

	// ErrBillNotExpanded is returned when trying to collapse a bill that is not expanded.
	ErrBillNotExpanded = errors.New("bill is not expanded")

	// ErrBillAlreadyExpanded is returned when trying to expand an already expanded bill.
	ErrBillAlreadyExpanded = errors.New("bill is already expanded")

	// ErrNoPotentialMatches is returned when no potential bill matches are found.
	ErrNoPotentialMatches = errors.New("no potential bill payment matches found")

	// ErrEmptyCCTransactions is returned when no CC transactions are provided.
	ErrEmptyCCTransactions = errors.New("credit card transactions list cannot be empty")

	// ErrBillPaymentNotOwned is returned when bill payment does not belong to user.
	ErrBillPaymentNotOwned = errors.New("bill payment does not belong to user")

	// Reconciliation errors.

	// ErrPendingNotFound is returned when no pending CC transactions exist for a billing cycle.
	ErrPendingNotFound = errors.New("no pending CC transactions for billing cycle")

	// ErrCycleAlreadyLinked is returned when trying to link a cycle that is already linked.
	ErrCycleAlreadyLinked = errors.New("billing cycle already has linked bill")

	// ErrAmountMismatch is returned when amount difference exceeds tolerance without force.
	ErrAmountMismatch = errors.New("amount difference exceeds tolerance")
)

// TransactionErrorCode defines error codes for transaction errors.
// Format: TXN-XXYYYY where XX is category and YYYY is specific error.
type TransactionErrorCode string

const (
	// Validation errors (01XXXX)
	ErrCodeInvalidTransactionType   TransactionErrorCode = "TXN-010001"
	ErrCodeInvalidTransactionDate   TransactionErrorCode = "TXN-010002"
	ErrCodeInvalidTransactionAmount TransactionErrorCode = "TXN-010003"
	ErrCodeTransactionNotFound      TransactionErrorCode = "TXN-010004"
	ErrCodeNotAuthorizedTransaction TransactionErrorCode = "TXN-010005"
	ErrCodeTxnCategoryNotFound      TransactionErrorCode = "TXN-010006"
	ErrCodeTxnCategoryNotOwned      TransactionErrorCode = "TXN-010007"
	ErrCodeDescriptionTooLong       TransactionErrorCode = "TXN-010008"
	ErrCodeNotesTooLong             TransactionErrorCode = "TXN-010009"
	ErrCodeMissingTransactionFields TransactionErrorCode = "TXN-010010"
	ErrCodeEmptyTransactionIDs      TransactionErrorCode = "TXN-010011"
	ErrCodeTransactionIDsNotFound   TransactionErrorCode = "TXN-010012"

	// Credit card import errors (02XXXX)
	ErrCodeInvalidBillingCycle  TransactionErrorCode = "TXN-020001"
	ErrCodeBillPaymentNotFound  TransactionErrorCode = "TXN-020002"
	ErrCodeBillNotExpanded      TransactionErrorCode = "TXN-020003"
	ErrCodeBillAlreadyExpanded  TransactionErrorCode = "TXN-020004"
	ErrCodeNoPotentialMatches   TransactionErrorCode = "TXN-020005"
	ErrCodeEmptyCCTransactions  TransactionErrorCode = "TXN-020006"
	ErrCodeBillPaymentNotOwned  TransactionErrorCode = "TXN-020007"

	// Reconciliation errors (03XXXX)
	ErrCodePendingNotFound    TransactionErrorCode = "TXN-030001"
	ErrCodeCycleAlreadyLinked TransactionErrorCode = "TXN-030002"
	ErrCodeAmountMismatch     TransactionErrorCode = "TXN-030003"

	// Internal errors (99XXXX)
	ErrCodeInternalError TransactionErrorCode = "TXN-990001"
)

// TransactionError represents a transaction error with code and message.
type TransactionError struct {
	Code    TransactionErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *TransactionError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *TransactionError) Unwrap() error {
	return e.Err
}

// NewTransactionError creates a new TransactionError with the given code and message.
func NewTransactionError(code TransactionErrorCode, message string, err error) *TransactionError {
	return &TransactionError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

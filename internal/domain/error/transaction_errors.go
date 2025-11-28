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
)

// TransactionErrorCode defines error codes for transaction errors.
// Format: TXN-XXYYYY where XX is category and YYYY is specific error.
type TransactionErrorCode string

const (
	// Validation errors (01XXXX)
	ErrCodeInvalidTransactionType       TransactionErrorCode = "TXN-010001"
	ErrCodeInvalidTransactionDate       TransactionErrorCode = "TXN-010002"
	ErrCodeInvalidTransactionAmount     TransactionErrorCode = "TXN-010003"
	ErrCodeTransactionNotFound          TransactionErrorCode = "TXN-010004"
	ErrCodeNotAuthorizedTransaction     TransactionErrorCode = "TXN-010005"
	ErrCodeTxnCategoryNotFound          TransactionErrorCode = "TXN-010006"
	ErrCodeTxnCategoryNotOwned          TransactionErrorCode = "TXN-010007"
	ErrCodeDescriptionTooLong           TransactionErrorCode = "TXN-010008"
	ErrCodeNotesTooLong                 TransactionErrorCode = "TXN-010009"
	ErrCodeMissingTransactionFields     TransactionErrorCode = "TXN-010010"
	ErrCodeEmptyTransactionIDs          TransactionErrorCode = "TXN-010011"
	ErrCodeTransactionIDsNotFound       TransactionErrorCode = "TXN-010012"
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

// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// CategoryRule domain errors.
var (
	// ErrCategoryRuleNotFound is returned when a category rule is not found in the system.
	ErrCategoryRuleNotFound = errors.New("category rule not found")

	// ErrCategoryRulePatternExists is returned when attempting to create a rule with an existing pattern.
	ErrCategoryRulePatternExists = errors.New("category rule pattern already exists")

	// ErrInvalidPattern is returned when the regex pattern is invalid.
	ErrInvalidPattern = errors.New("invalid regex pattern")

	// ErrPatternTooLong is returned when the pattern exceeds the maximum length.
	ErrPatternTooLong = errors.New("pattern too long")

	// ErrNotAuthorizedToModifyRule is returned when user is not authorized to modify a rule.
	ErrNotAuthorizedToModifyRule = errors.New("not authorized to modify rule")

	// ErrCategoryRuleMissingFields is returned when required fields are missing.
	ErrCategoryRuleMissingFields = errors.New("missing required fields")

	// ErrInvalidPriority is returned when the priority value is invalid.
	ErrInvalidPriority = errors.New("invalid priority value")
)

// CategoryRuleErrorCode defines error codes for category rule errors.
// Format: CRL-XXYYYY where XX is category and YYYY is specific error.
type CategoryRuleErrorCode string

const (
	// Validation errors (01XXXX)
	ErrCodeCategoryRuleNotFound      CategoryRuleErrorCode = "CRL-010001"
	ErrCodeCategoryRulePatternExists CategoryRuleErrorCode = "CRL-010002"
	ErrCodeInvalidPattern            CategoryRuleErrorCode = "CRL-010003"
	ErrCodePatternTooLong            CategoryRuleErrorCode = "CRL-010004"
	ErrCodeNotAuthorizedRule         CategoryRuleErrorCode = "CRL-010005"
	ErrCodeMissingRuleFields         CategoryRuleErrorCode = "CRL-010006"
	ErrCodeInvalidPriority           CategoryRuleErrorCode = "CRL-010007"
	ErrCodeCategoryNotFoundForRule   CategoryRuleErrorCode = "CRL-010008"
	ErrCodeRuleOwnerTypeMismatch     CategoryRuleErrorCode = "CRL-010009"
)

// CategoryRuleError represents a category rule error with code and message.
type CategoryRuleError struct {
	Code    CategoryRuleErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *CategoryRuleError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *CategoryRuleError) Unwrap() error {
	return e.Err
}

// NewCategoryRuleError creates a new CategoryRuleError with the given code and message.
func NewCategoryRuleError(code CategoryRuleErrorCode, message string, err error) *CategoryRuleError {
	return &CategoryRuleError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

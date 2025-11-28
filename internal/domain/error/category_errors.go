// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// Category domain errors.
var (
	// ErrCategoryNotFound is returned when a category is not found in the system.
	ErrCategoryNotFound = errors.New("category not found")

	// ErrCategoryNameExists is returned when attempting to create a category with an existing name.
	ErrCategoryNameExists = errors.New("category name already exists")

	// ErrCategoryNameTooLong is returned when the category name exceeds the maximum length.
	ErrCategoryNameTooLong = errors.New("category name too long")

	// ErrInvalidColorFormat is returned when the category color format is invalid.
	ErrInvalidColorFormat = errors.New("invalid color format")

	// ErrInvalidOwnerType is returned when the owner type is invalid.
	ErrInvalidOwnerType = errors.New("invalid owner type")

	// ErrNotAuthorizedToModifyCategory is returned when user is not authorized to modify a category.
	ErrNotAuthorizedToModifyCategory = errors.New("not authorized to modify category")

	// ErrInvalidCategoryType is returned when the category type is invalid.
	ErrInvalidCategoryType = errors.New("invalid category type")
)

// CategoryErrorCode defines error codes for category errors.
// Format: CAT-XXYYYY where XX is category and YYYY is specific error.
type CategoryErrorCode string

const (
	// Validation errors (01XXXX)
	ErrCodeCategoryNameTooLong   CategoryErrorCode = "CAT-010001"
	ErrCodeInvalidColorFormat    CategoryErrorCode = "CAT-010002"
	ErrCodeInvalidOwnerType      CategoryErrorCode = "CAT-010003"
	ErrCodeCategoryNotFound      CategoryErrorCode = "CAT-010004"
	ErrCodeCategoryNameExists    CategoryErrorCode = "CAT-010005"
	ErrCodeNotAuthorizedCategory CategoryErrorCode = "CAT-010006"
	ErrCodeInvalidCategoryType   CategoryErrorCode = "CAT-010007"
	ErrCodeMissingCategoryFields CategoryErrorCode = "CAT-010008"
)

// CategoryError represents a category error with code and message.
type CategoryError struct {
	Code    CategoryErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *CategoryError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *CategoryError) Unwrap() error {
	return e.Err
}

// NewCategoryError creates a new CategoryError with the given code and message.
func NewCategoryError(code CategoryErrorCode, message string, err error) *CategoryError {
	return &CategoryError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

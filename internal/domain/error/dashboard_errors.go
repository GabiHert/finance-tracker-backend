// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// Dashboard domain errors.
var (
	// ErrMissingStartDate is returned when start_date is not provided.
	ErrMissingStartDate = errors.New("start_date is required")

	// ErrMissingEndDate is returned when end_date is not provided.
	ErrMissingEndDate = errors.New("end_date is required")

	// ErrInvalidDateRange is returned when end_date is before start_date.
	ErrInvalidDateRange = errors.New("end_date must be after start_date")

	// ErrInvalidGranularity is returned when granularity is not valid.
	ErrInvalidGranularity = errors.New("granularity must be: daily, weekly, or monthly")

	// ErrMissingGranularity is returned when granularity is not provided.
	ErrMissingGranularity = errors.New("granularity is required")

	// ErrInvalidDateFormat is returned when date format is invalid.
	ErrInvalidDateFormat = errors.New("invalid date format, expected YYYY-MM-DD")
)

// DashboardErrorCode defines error codes for dashboard errors.
// Format: DSH-XXYYYY where XX is category and YYYY is specific error.
type DashboardErrorCode string

const (
	// Validation errors (01XXXX)
	ErrCodeMissingStartDate   DashboardErrorCode = "DSH-010001"
	ErrCodeMissingEndDate     DashboardErrorCode = "DSH-010002"
	ErrCodeInvalidDateRange   DashboardErrorCode = "DSH-010003"
	ErrCodeInvalidGranularity DashboardErrorCode = "DSH-010004"
	ErrCodeMissingGranularity DashboardErrorCode = "DSH-010005"
	ErrCodeInvalidDateFormat  DashboardErrorCode = "DSH-010006"

	// Internal errors (99XXXX)
	ErrCodeDashboardInternalError DashboardErrorCode = "DSH-990001"
)

// DashboardError represents a dashboard error with code and message.
type DashboardError struct {
	Code    DashboardErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *DashboardError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *DashboardError) Unwrap() error {
	return e.Err
}

// NewDashboardError creates a new DashboardError with the given code and message.
func NewDashboardError(code DashboardErrorCode, message string, err error) *DashboardError {
	return &DashboardError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// Email domain errors.
var (
	// ErrEmailQueueFailed is returned when an email fails to be queued.
	ErrEmailQueueFailed = errors.New("failed to queue email")

	// ErrEmailSendFailed is returned when an email fails to be sent.
	ErrEmailSendFailed = errors.New("failed to send email")

	// ErrInvalidTemplate is returned when an invalid email template is specified.
	ErrInvalidTemplate = errors.New("invalid email template")

	// ErrTemplateRenderFailed is returned when email template rendering fails.
	ErrTemplateRenderFailed = errors.New("failed to render email template")

	// ErrEmailJobNotFound is returned when an email job is not found.
	ErrEmailJobNotFound = errors.New("email job not found")

	// ErrPermanentEmailFailure is returned when an email fails with a permanent error.
	ErrPermanentEmailFailure = errors.New("permanent email failure")

	// ErrTemporaryEmailFailure is returned when an email fails with a temporary error.
	ErrTemporaryEmailFailure = errors.New("temporary email failure")
)

// EmailErrorCode defines error codes for email errors.
// Format: EMAIL-XXYYYY where XX is category and YYYY is specific error.
type EmailErrorCode string

const (
	// Queue errors (01XXXX)
	ErrCodeEmailQueueFailed EmailErrorCode = "EMAIL-010001"
	ErrCodeEmailJobNotFound EmailErrorCode = "EMAIL-010002"

	// Send errors (02XXXX)
	ErrCodeEmailSendFailed       EmailErrorCode = "EMAIL-020001"
	ErrCodePermanentEmailFailure EmailErrorCode = "EMAIL-020002"
	ErrCodeTemporaryEmailFailure EmailErrorCode = "EMAIL-020003"

	// Template errors (03XXXX)
	ErrCodeInvalidTemplate      EmailErrorCode = "EMAIL-030001"
	ErrCodeTemplateRenderFailed EmailErrorCode = "EMAIL-030002"
)

// EmailError represents an email error with code and message.
type EmailError struct {
	Code    EmailErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *EmailError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *EmailError) Unwrap() error {
	return e.Err
}

// NewEmailError creates a new EmailError with the given code and message.
func NewEmailError(code EmailErrorCode, message string, err error) *EmailError {
	return &EmailError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// Authentication domain errors.
var (
	// ErrUserNotFound is returned when a user is not found in the system.
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailAlreadyExists is returned when attempting to register with an existing email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidToken is returned when a token is invalid or malformed.
	ErrInvalidToken = errors.New("invalid token")

	// ErrExpiredToken is returned when a token has expired.
	ErrExpiredToken = errors.New("token has expired")

	// ErrInvalidResetToken is returned when a password reset token is invalid.
	ErrInvalidResetToken = errors.New("invalid or expired password reset token")

	// ErrTermsNotAccepted is returned when user has not accepted the terms of service.
	ErrTermsNotAccepted = errors.New("terms of service must be accepted")

	// ErrWeakPassword is returned when the provided password does not meet requirements.
	ErrWeakPassword = errors.New("password does not meet minimum requirements")

	// ErrInvalidEmail is returned when the provided email format is invalid.
	ErrInvalidEmail = errors.New("invalid email format")
)

// AuthErrorCode defines error codes for authentication errors.
// Format: AUTH-XXYYYY where XX is category and YYYY is specific error.
type AuthErrorCode string

const (
	// Registration errors (01XXXX)
	ErrCodeEmailExists       AuthErrorCode = "AUTH-010001"
	ErrCodeTermsNotAccepted  AuthErrorCode = "AUTH-010002"
	ErrCodeWeakPassword      AuthErrorCode = "AUTH-010003"
	ErrCodeInvalidEmail      AuthErrorCode = "AUTH-010004"
	ErrCodeMissingFields     AuthErrorCode = "AUTH-010005"

	// Login errors (02XXXX)
	ErrCodeInvalidCredentials AuthErrorCode = "AUTH-020001"
	ErrCodeUserNotFound       AuthErrorCode = "AUTH-020002"
	ErrCodeRateLimited        AuthErrorCode = "AUTH-020003"

	// Token errors (03XXXX)
	ErrCodeInvalidToken       AuthErrorCode = "AUTH-030001"
	ErrCodeExpiredToken       AuthErrorCode = "AUTH-030002"
	ErrCodeMissingToken       AuthErrorCode = "AUTH-030003"

	// Password reset errors (04XXXX)
	ErrCodeInvalidResetToken  AuthErrorCode = "AUTH-040001"
	ErrCodeExpiredResetToken  AuthErrorCode = "AUTH-040002"

	// Delete account errors (05XXXX)
	ErrCodeInvalidConfirmation AuthErrorCode = "AUTH-050001"
)

// AuthError represents an authentication error with code and message.
type AuthError struct {
	Code    AuthErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewAuthError creates a new AuthError with the given code and message.
func NewAuthError(code AuthErrorCode, message string, err error) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

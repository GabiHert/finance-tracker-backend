// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// Goal domain errors.
var (
	// ErrGoalNotFound is returned when a goal is not found in the system.
	ErrGoalNotFound = errors.New("goal not found")

	// ErrGoalAlreadyExists is returned when attempting to create a goal for a category that already has one.
	ErrGoalAlreadyExists = errors.New("goal already exists for this category")

	// ErrInvalidLimitAmount is returned when the limit amount is invalid (zero or negative).
	ErrInvalidLimitAmount = errors.New("invalid limit amount")

	// ErrGoalCategoryNotFound is returned when the category for a goal is not found.
	ErrGoalCategoryNotFound = errors.New("category not found")

	// ErrCategoryDoesNotBelongToUser is returned when the category does not belong to the user.
	ErrCategoryDoesNotBelongToUser = errors.New("category does not belong to user")

	// ErrUnauthorizedGoalAccess is returned when user is not authorized to access a goal.
	ErrUnauthorizedGoalAccess = errors.New("unauthorized access to goal")

	// ErrInvalidGoalPeriod is returned when the goal period is invalid.
	ErrInvalidGoalPeriod = errors.New("invalid goal period")
)

// GoalErrorCode defines error codes for goal errors.
// Format: GOL-XXYYYY where XX is category and YYYY is specific error.
type GoalErrorCode string

const (
	// Validation errors (01XXXX)
	ErrCodeGoalNotFound              GoalErrorCode = "GOL-010001"
	ErrCodeGoalAlreadyExists         GoalErrorCode = "GOL-010002"
	ErrCodeInvalidLimitAmount        GoalErrorCode = "GOL-010003"
	ErrCodeGoalCategoryNotFound      GoalErrorCode = "GOL-010004"
	ErrCodeCategoryDoesNotBelongUser GoalErrorCode = "GOL-010005"
	ErrCodeUnauthorizedGoalAccess    GoalErrorCode = "GOL-010006"
	ErrCodeInvalidGoalPeriod         GoalErrorCode = "GOL-010007"
	ErrCodeMissingGoalFields         GoalErrorCode = "GOL-010008"
)

// GoalError represents a goal error with code and message.
type GoalError struct {
	Code    GoalErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *GoalError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *GoalError) Unwrap() error {
	return e.Err
}

// NewGoalError creates a new GoalError with the given code and message.
func NewGoalError(code GoalErrorCode, message string, err error) *GoalError {
	return &GoalError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

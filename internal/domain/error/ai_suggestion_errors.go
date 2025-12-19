// Package error defines domain-specific errors for the Finance Tracker application.
package error

import "errors"

// AI Categorization domain errors.
var (
	// ErrAISuggestionNotFound is returned when an AI suggestion is not found in the system.
	ErrAISuggestionNotFound = errors.New("ai suggestion not found")

	// ErrAIAlreadyProcessing is returned when attempting to start categorization while already processing.
	ErrAIAlreadyProcessing = errors.New("ai categorization already in progress")

	// ErrAIPatternConflict is returned when the suggested pattern conflicts with an existing rule.
	ErrAIPatternConflict = errors.New("pattern conflicts with existing rule")

	// ErrAINoUncategorized is returned when there are no uncategorized transactions to process.
	ErrAINoUncategorized = errors.New("no uncategorized transactions found")

	// ErrAIRetryFailed is returned when retrying with AI fails.
	ErrAIRetryFailed = errors.New("ai retry failed")

	// ErrAIServiceError is returned when the AI service encounters an error.
	ErrAIServiceError = errors.New("ai service error")

	// ErrAIRateLimited is returned when the AI service rate limits requests.
	ErrAIRateLimited = errors.New("ai service rate limited")

	// ErrAIInvalidMatchType is returned when the match type is invalid.
	ErrAIInvalidMatchType = errors.New("invalid match type")

	// ErrAIEmptyKeyword is returned when the match keyword is empty.
	ErrAIEmptyKeyword = errors.New("match keyword cannot be empty")

	// ErrAISuggestionAlreadyProcessed is returned when trying to process an already processed suggestion.
	ErrAISuggestionAlreadyProcessed = errors.New("suggestion has already been processed")

	// ErrAIInvalidAction is returned when an invalid action is provided.
	ErrAIInvalidAction = errors.New("invalid action")
)

// AISuggestionErrorCode defines error codes for AI categorization errors.
// Format: AIC-XXYYYY where XX is category and YYYY is specific error.
type AISuggestionErrorCode string

const (
	// Validation errors (01XXXX)
	ErrCodeAISuggestionNotFound         AISuggestionErrorCode = "AIC-010001"
	ErrCodeAIAlreadyProcessing          AISuggestionErrorCode = "AIC-010002"
	ErrCodeAINoUncategorized            AISuggestionErrorCode = "AIC-010003"
	ErrCodeAIPatternConflict            AISuggestionErrorCode = "AIC-010004"
	ErrCodeAIInvalidMatchType           AISuggestionErrorCode = "AIC-010005"
	ErrCodeAIEmptyKeyword               AISuggestionErrorCode = "AIC-010006"
	ErrCodeAISuggestionAlreadyProcessed AISuggestionErrorCode = "AIC-010007"
	ErrCodeAIInvalidAction              AISuggestionErrorCode = "AIC-010008"

	// External service errors (02XXXX)
	ErrCodeAIServiceError  AISuggestionErrorCode = "AIC-020001"
	ErrCodeAIRateLimited   AISuggestionErrorCode = "AIC-020002"
	ErrCodeAIRetryFailed   AISuggestionErrorCode = "AIC-020003"
	ErrCodeAIInvalidConfig AISuggestionErrorCode = "AIC-020004"
)

// AISuggestionError represents an AI categorization error with code and message.
type AISuggestionError struct {
	Code    AISuggestionErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *AISuggestionError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *AISuggestionError) Unwrap() error {
	return e.Err
}

// NewAISuggestionError creates a new AISuggestionError with the given code and message.
func NewAISuggestionError(code AISuggestionErrorCode, message string, err error) *AISuggestionError {
	return &AISuggestionError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

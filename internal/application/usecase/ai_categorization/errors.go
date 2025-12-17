// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Error code constants for AI processing errors.
const (
	ErrCodeAIServiceUnavailable = "AI_SERVICE_UNAVAILABLE"
	ErrCodeAIRateLimited        = "AI_RATE_LIMITED"
	ErrCodeAIAuthError          = "AI_AUTH_ERROR"
	ErrCodeAITimeout            = "AI_TIMEOUT"
	ErrCodeAIParseError         = "AI_PARSE_ERROR"
	ErrCodeAIUnknownError       = "AI_UNKNOWN_ERROR"
)

// errorMessages contains Portuguese error messages for each error code.
var errorMessages = map[string]string{
	ErrCodeAIServiceUnavailable: "O servico de inteligencia artificial esta temporariamente indisponivel. Tente novamente mais tarde.",
	ErrCodeAIRateLimited:        "Limite de requisicoes atingido. Aguarde alguns minutos e tente novamente.",
	ErrCodeAIAuthError:          "Erro de configuracao do servico de IA. Por favor, contate o suporte.",
	ErrCodeAITimeout:            "O processamento demorou mais do que o esperado. Tente novamente com menos transacoes.",
	ErrCodeAIParseError:         "Erro ao processar resposta da IA. Tente novamente.",
	ErrCodeAIUnknownError:       "Ocorreu um erro inesperado durante o processamento. Tente novamente.",
}

// ProcessingError represents an error that occurred during AI processing.
type ProcessingError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Retryable bool      `json:"retryable"`
	Timestamp time.Time `json:"timestamp"`
}

// classifyError converts a Go error to a ProcessingError with appropriate
// error code, Portuguese message, and retryable flag.
func classifyError(err error) *ProcessingError {
	now := time.Now()
	errStr := strings.ToLower(err.Error())

	// Check for timeout/cancellation (context errors)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return &ProcessingError{
			Code:      ErrCodeAITimeout,
			Message:   errorMessages[ErrCodeAITimeout],
			Retryable: true,
			Timestamp: now,
		}
	}

	// Check for rate limiting
	if strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "quota") ||
		strings.Contains(errStr, "429") || strings.Contains(errStr, "resource exhausted") {
		return &ProcessingError{
			Code:      ErrCodeAIRateLimited,
			Message:   errorMessages[ErrCodeAIRateLimited],
			Retryable: true,
			Timestamp: now,
		}
	}

	// Check for authentication errors
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "invalid api key") || strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "authentication") {
		return &ProcessingError{
			Code:      ErrCodeAIAuthError,
			Message:   errorMessages[ErrCodeAIAuthError],
			Retryable: false,
			Timestamp: now,
		}
	}

	// Check for network/connection errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "dial") || strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "503") {
		return &ProcessingError{
			Code:      ErrCodeAIServiceUnavailable,
			Message:   errorMessages[ErrCodeAIServiceUnavailable],
			Retryable: true,
			Timestamp: now,
		}
	}

	// Check for parse errors
	if strings.Contains(errStr, "parse") || strings.Contains(errStr, "json") ||
		strings.Contains(errStr, "unmarshal") || strings.Contains(errStr, "decode") {
		return &ProcessingError{
			Code:      ErrCodeAIParseError,
			Message:   errorMessages[ErrCodeAIParseError],
			Retryable: true,
			Timestamp: now,
		}
	}

	// Default to unknown error
	return &ProcessingError{
		Code:      ErrCodeAIUnknownError,
		Message:   errorMessages[ErrCodeAIUnknownError],
		Retryable: true,
		Timestamp: now,
	}
}

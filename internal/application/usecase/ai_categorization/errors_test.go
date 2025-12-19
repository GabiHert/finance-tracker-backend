// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"
	"errors"
	"testing"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
		expectRetry  bool
	}{
		// Timeout/cancellation errors
		{
			name:         "context deadline exceeded",
			err:          context.DeadlineExceeded,
			expectedCode: ErrCodeAITimeout,
			expectRetry:  true,
		},
		{
			name:         "context canceled",
			err:          context.Canceled,
			expectedCode: ErrCodeAITimeout,
			expectRetry:  true,
		},
		// Rate limiting errors
		{
			name:         "rate limit error",
			err:          errors.New("rate limit exceeded"),
			expectedCode: ErrCodeAIRateLimited,
			expectRetry:  true,
		},
		{
			name:         "quota error",
			err:          errors.New("quota exceeded"),
			expectedCode: ErrCodeAIRateLimited,
			expectRetry:  true,
		},
		{
			name:         "429 status code error",
			err:          errors.New("HTTP 429: too many requests"),
			expectedCode: ErrCodeAIRateLimited,
			expectRetry:  true,
		},
		{
			name:         "resource exhausted error",
			err:          errors.New("resource exhausted"),
			expectedCode: ErrCodeAIRateLimited,
			expectRetry:  true,
		},
		// Authentication errors
		{
			name:         "401 unauthorized",
			err:          errors.New("401 unauthorized"),
			expectedCode: ErrCodeAIAuthError,
			expectRetry:  false,
		},
		{
			name:         "403 forbidden",
			err:          errors.New("403 forbidden"),
			expectedCode: ErrCodeAIAuthError,
			expectRetry:  false,
		},
		{
			name:         "invalid api key",
			err:          errors.New("invalid api key"),
			expectedCode: ErrCodeAIAuthError,
			expectRetry:  false,
		},
		{
			name:         "unauthorized error",
			err:          errors.New("unauthorized access"),
			expectedCode: ErrCodeAIAuthError,
			expectRetry:  false,
		},
		{
			name:         "authentication error",
			err:          errors.New("authentication failed"),
			expectedCode: ErrCodeAIAuthError,
			expectRetry:  false,
		},
		// Network/connection errors
		{
			name:         "connection refused",
			err:          errors.New("connection refused"),
			expectedCode: ErrCodeAIServiceUnavailable,
			expectRetry:  true,
		},
		{
			name:         "network error",
			err:          errors.New("network error"),
			expectedCode: ErrCodeAIServiceUnavailable,
			expectRetry:  true,
		},
		{
			name:         "dial error",
			err:          errors.New("dial tcp: connection refused"),
			expectedCode: ErrCodeAIServiceUnavailable,
			expectRetry:  true,
		},
		{
			name:         "timeout error in message",
			err:          errors.New("request timeout"),
			expectedCode: ErrCodeAIServiceUnavailable,
			expectRetry:  true,
		},
		{
			name:         "unavailable error",
			err:          errors.New("service unavailable"),
			expectedCode: ErrCodeAIServiceUnavailable,
			expectRetry:  true,
		},
		{
			name:         "503 status code error",
			err:          errors.New("HTTP 503: service unavailable"),
			expectedCode: ErrCodeAIServiceUnavailable,
			expectRetry:  true,
		},
		// Parse errors
		{
			name:         "parse error",
			err:          errors.New("failed to parse response"),
			expectedCode: ErrCodeAIParseError,
			expectRetry:  true,
		},
		{
			name:         "json error",
			err:          errors.New("invalid json"),
			expectedCode: ErrCodeAIParseError,
			expectRetry:  true,
		},
		{
			name:         "unmarshal error",
			err:          errors.New("unmarshal failed"),
			expectedCode: ErrCodeAIParseError,
			expectRetry:  true,
		},
		{
			name:         "decode error",
			err:          errors.New("decode error"),
			expectedCode: ErrCodeAIParseError,
			expectRetry:  true,
		},
		// Unknown errors
		{
			name:         "unknown error",
			err:          errors.New("something unexpected happened"),
			expectedCode: ErrCodeAIUnknownError,
			expectRetry:  true,
		},
		{
			name:         "generic error",
			err:          errors.New("internal server error"),
			expectedCode: ErrCodeAIUnknownError,
			expectRetry:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyError(tt.err)

			if result.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, result.Code)
			}

			if result.Retryable != tt.expectRetry {
				t.Errorf("expected retryable %v, got %v", tt.expectRetry, result.Retryable)
			}

			if result.Message == "" {
				t.Error("expected non-empty message")
			}

			if result.Timestamp.IsZero() {
				t.Error("expected non-zero timestamp")
			}

			// Verify the message is the expected Portuguese message
			expectedMessage := errorMessages[tt.expectedCode]
			if result.Message != expectedMessage {
				t.Errorf("expected message %q, got %q", expectedMessage, result.Message)
			}
		})
	}
}

func TestClassifyError_WrappedContextErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "wrapped deadline exceeded",
			err:          context.DeadlineExceeded,
			expectedCode: ErrCodeAITimeout,
		},
		{
			name:         "wrapped canceled",
			err:          context.Canceled,
			expectedCode: ErrCodeAITimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyError(tt.err)

			if result.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, result.Code)
			}
		})
	}
}

func TestClassifyError_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "uppercase rate limit",
			err:          errors.New("RATE LIMIT exceeded"),
			expectedCode: ErrCodeAIRateLimited,
		},
		{
			name:         "mixed case network",
			err:          errors.New("Network Error"),
			expectedCode: ErrCodeAIServiceUnavailable,
		},
		{
			name:         "uppercase json",
			err:          errors.New("Invalid JSON format"),
			expectedCode: ErrCodeAIParseError,
		},
		{
			name:         "mixed case unauthorized",
			err:          errors.New("Unauthorized access"),
			expectedCode: ErrCodeAIAuthError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyError(tt.err)

			if result.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, result.Code)
			}
		})
	}
}

func TestErrorMessages_AllCodesHaveMessages(t *testing.T) {
	codes := []string{
		ErrCodeAIServiceUnavailable,
		ErrCodeAIRateLimited,
		ErrCodeAIAuthError,
		ErrCodeAITimeout,
		ErrCodeAIParseError,
		ErrCodeAIUnknownError,
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			message, exists := errorMessages[code]
			if !exists {
				t.Errorf("missing message for code %s", code)
			}
			if message == "" {
				t.Errorf("empty message for code %s", code)
			}
		})
	}
}

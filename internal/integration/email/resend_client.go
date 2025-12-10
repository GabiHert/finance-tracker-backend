// Package email provides email sending functionality via Resend.
package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/resend/resend-go/v2"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// ResendClient implements the adapter.EmailSender interface using Resend.
type ResendClient struct {
	client    *resend.Client
	fromName  string
	fromEmail string
}

// NewResendClient creates a new Resend client.
func NewResendClient(apiKey, fromName, fromEmail string) *ResendClient {
	return &ResendClient{
		client:    resend.NewClient(apiKey),
		fromName:  fromName,
		fromEmail: fromEmail,
	}
}

// Send sends an email via Resend.
func (c *ResendClient) Send(ctx context.Context, input adapter.SendEmailInput) (*adapter.SendEmailResult, error) {
	from := fmt.Sprintf("%s <%s>", c.fromName, c.fromEmail)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{input.To},
		Subject: input.Subject,
		Html:    input.HTML,
		Text:    input.Text,
	}

	resp, err := c.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		// Check if it's a permanent error (don't retry)
		if isPermanentError(err) {
			return nil, domainerror.NewEmailError(
				domainerror.ErrCodePermanentEmailFailure,
				"permanent email failure",
				err,
			)
		}
		// Temporary error (can retry)
		return nil, domainerror.NewEmailError(
			domainerror.ErrCodeTemporaryEmailFailure,
			"temporary email failure",
			err,
		)
	}

	return &adapter.SendEmailResult{
		ResendID: resp.Id,
	}, nil
}

// isPermanentError checks if the error is a permanent error that should not be retried.
// Permanent errors include: 401 (Unauthorized), 403 (Forbidden), 422 (Validation Error)
// Temporary errors include: 429 (Rate Limit), 5xx (Server Errors)
func isPermanentError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check for common permanent error patterns
	permanentPatterns := []string{
		"401",
		"403",
		"422",
		"unauthorized",
		"forbidden",
		"validation",
		"invalid",
		"bad request",
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// MockEmailSender is a mock implementation for testing.
type MockEmailSender struct {
	SentEmails  []adapter.SendEmailInput
	ShouldFail  bool
	FailError   error
	IsPermanent bool
}

// NewMockEmailSender creates a new mock email sender.
func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		SentEmails: make([]adapter.SendEmailInput, 0),
	}
}

// Send implements the adapter.EmailSender interface for testing.
func (m *MockEmailSender) Send(ctx context.Context, input adapter.SendEmailInput) (*adapter.SendEmailResult, error) {
	if m.ShouldFail {
		if m.IsPermanent {
			return nil, domainerror.NewEmailError(
				domainerror.ErrCodePermanentEmailFailure,
				"mock permanent failure",
				m.FailError,
			)
		}
		return nil, domainerror.NewEmailError(
			domainerror.ErrCodeTemporaryEmailFailure,
			"mock temporary failure",
			m.FailError,
		)
	}

	m.SentEmails = append(m.SentEmails, input)

	return &adapter.SendEmailResult{
		ResendID: fmt.Sprintf("mock-%d", len(m.SentEmails)),
	}, nil
}

// SetFailure configures the mock to fail with the given error.
func (m *MockEmailSender) SetFailure(err error, permanent bool) {
	m.ShouldFail = true
	m.FailError = err
	m.IsPermanent = permanent
}

// ClearFailure clears the failure configuration.
func (m *MockEmailSender) ClearFailure() {
	m.ShouldFail = false
	m.FailError = nil
	m.IsPermanent = false
}

// Reset clears all sent emails and failure configuration.
func (m *MockEmailSender) Reset() {
	m.SentEmails = make([]adapter.SendEmailInput, 0)
	m.ClearFailure()
}

// Ensure implementations satisfy interfaces.
var (
	_ adapter.EmailSender = (*ResendClient)(nil)
	_ adapter.EmailSender = (*MockEmailSender)(nil)
)

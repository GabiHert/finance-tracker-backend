// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"
)

// SendEmailInput represents the input for sending an email.
type SendEmailInput struct {
	To      string
	Name    string
	Subject string
	HTML    string
	Text    string
}

// SendEmailResult represents the result of sending an email.
type SendEmailResult struct {
	ResendID string
}

// EmailSender defines the interface for sending emails via an external provider.
type EmailSender interface {
	// Send sends an email via the email provider (e.g., Resend).
	Send(ctx context.Context, input SendEmailInput) (*SendEmailResult, error)
}

// EmailService defines the interface for queueing emails.
type EmailService interface {
	// QueuePasswordResetEmail queues a password reset email.
	QueuePasswordResetEmail(ctx context.Context, input QueuePasswordResetInput) error

	// QueueGroupInvitationEmail queues a group invitation email.
	QueueGroupInvitationEmail(ctx context.Context, input QueueGroupInvitationInput) error
}

// QueuePasswordResetInput represents the input for queueing a password reset email.
type QueuePasswordResetInput struct {
	UserID    string
	UserEmail string
	UserName  string
	ResetURL  string
	ExpiresIn string
}

// QueueGroupInvitationInput represents the input for queueing a group invitation email.
type QueueGroupInvitationInput struct {
	InviterName  string
	InviterEmail string
	GroupName    string
	InviteEmail  string
	InviteURL    string
	ExpiresIn    string
}

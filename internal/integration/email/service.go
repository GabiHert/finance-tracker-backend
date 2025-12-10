// Package email provides email sending functionality.
package email

import (
	"context"
	"fmt"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// Service handles email queueing operations.
type Service struct {
	queue      adapter.EmailQueueRepository
	appBaseURL string
}

// NewService creates a new email service.
func NewService(queue adapter.EmailQueueRepository, appBaseURL string) *Service {
	return &Service{
		queue:      queue,
		appBaseURL: appBaseURL,
	}
}

// QueuePasswordResetEmail queues a password reset email.
func (s *Service) QueuePasswordResetEmail(ctx context.Context, input adapter.QueuePasswordResetInput) error {
	subject := "Redefinir sua senha - Finance Tracker"

	templateData := map[string]interface{}{
		"user_name":  input.UserName,
		"reset_url":  input.ResetURL,
		"expires_in": input.ExpiresIn,
	}

	job := entity.NewEmailJob(
		entity.TemplatePasswordReset,
		input.UserEmail,
		input.UserName,
		subject,
		templateData,
	)

	if err := s.queue.Create(ctx, job); err != nil {
		return domainerror.NewEmailError(
			domainerror.ErrCodeEmailQueueFailed,
			"failed to queue password reset email",
			err,
		)
	}

	return nil
}

// QueueGroupInvitationEmail queues a group invitation email.
func (s *Service) QueueGroupInvitationEmail(ctx context.Context, input adapter.QueueGroupInvitationInput) error {
	subject := fmt.Sprintf("%s convidou voce para %s - Finance Tracker", input.InviterName, input.GroupName)

	templateData := map[string]interface{}{
		"inviter_name":  input.InviterName,
		"inviter_email": input.InviterEmail,
		"group_name":    input.GroupName,
		"invite_url":    input.InviteURL,
		"expires_in":    input.ExpiresIn,
	}

	job := entity.NewEmailJob(
		entity.TemplateGroupInvitation,
		input.InviteEmail,
		"", // Recipient name unknown for invitations
		subject,
		templateData,
	)

	if err := s.queue.Create(ctx, job); err != nil {
		return domainerror.NewEmailError(
			domainerror.ErrCodeEmailQueueFailed,
			"failed to queue group invitation email",
			err,
		)
	}

	return nil
}

// Ensure Service implements adapter.EmailService.
var _ adapter.EmailService = (*Service)(nil)

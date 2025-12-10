// Package email provides email sending functionality.
package email

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/email/templates"
)

// Worker processes the email queue and sends emails.
type Worker struct {
	queue        adapter.EmailQueueRepository
	sender       adapter.EmailSender
	renderer     *templates.Renderer
	pollInterval time.Duration
	batchSize    int
}

// WorkerConfig holds configuration for the email worker.
type WorkerConfig struct {
	PollInterval time.Duration
	BatchSize    int
}

// DefaultWorkerConfig returns the default worker configuration.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		PollInterval: 5 * time.Second,
		BatchSize:    10,
	}
}

// NewWorker creates a new email worker.
func NewWorker(queue adapter.EmailQueueRepository, sender adapter.EmailSender, renderer *templates.Renderer, config WorkerConfig) *Worker {
	return &Worker{
		queue:        queue,
		sender:       sender,
		renderer:     renderer,
		pollInterval: config.PollInterval,
		batchSize:    config.BatchSize,
	}
}

// Start begins the worker loop. It blocks until the context is cancelled.
func (w *Worker) Start(ctx context.Context) {
	slog.Info("Email worker started",
		"poll_interval", w.pollInterval,
		"batch_size", w.batchSize,
	)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Process immediately on start, then on ticker
	w.processBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Email worker shutting down")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

// processBatch fetches and processes a batch of pending emails.
func (w *Worker) processBatch(ctx context.Context) {
	jobs, err := w.queue.GetPendingJobs(ctx, w.batchSize)
	if err != nil {
		slog.Error("Failed to get pending email jobs", "error", err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	slog.Debug("Processing email batch", "count", len(jobs))

	for _, job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			w.processJob(ctx, job)
		}
	}
}

// processJob processes a single email job.
func (w *Worker) processJob(ctx context.Context, job *entity.EmailJob) {
	logger := slog.With(
		"job_id", job.ID,
		"template", job.TemplateType,
		"recipient", job.RecipientEmail,
	)

	// Mark as processing
	job.MarkProcessing()
	if err := w.queue.Update(ctx, job); err != nil {
		logger.Error("Failed to mark job as processing", "error", err)
		return
	}

	// Render template
	html, text, err := w.renderTemplate(job)
	if err != nil {
		logger.Error("Failed to render email template", "error", err)
		w.handleFailure(ctx, job, err, true) // Template errors are permanent
		return
	}

	// Send email
	result, err := w.sender.Send(ctx, adapter.SendEmailInput{
		To:      job.RecipientEmail,
		Name:    job.RecipientName,
		Subject: job.Subject,
		HTML:    html,
		Text:    text,
	})

	if err != nil {
		logger.Error("Failed to send email", "error", err)

		// Check if it's a permanent error
		var emailErr *domainerror.EmailError
		isPermanent := errors.As(err, &emailErr) && emailErr.Code == domainerror.ErrCodePermanentEmailFailure

		w.handleFailure(ctx, job, err, isPermanent)
		return
	}

	// Mark as sent
	job.MarkSent(result.ResendID)
	if err := w.queue.Update(ctx, job); err != nil {
		logger.Error("Failed to mark job as sent", "error", err)
		return
	}

	logger.Info("Email sent successfully", "resend_id", result.ResendID)
}

// renderTemplate renders the appropriate template for the job.
func (w *Worker) renderTemplate(job *entity.EmailJob) (html string, text string, err error) {
	templateName := string(job.TemplateType)

	// Convert template data to the appropriate struct
	var data interface{}
	switch job.TemplateType {
	case entity.TemplatePasswordReset:
		data = templates.PasswordResetData{
			UserName:  getString(job.TemplateData, "user_name"),
			ResetURL:  getString(job.TemplateData, "reset_url"),
			ExpiresIn: getString(job.TemplateData, "expires_in"),
		}
	case entity.TemplateGroupInvitation:
		data = templates.GroupInvitationData{
			InviterName:  getString(job.TemplateData, "inviter_name"),
			InviterEmail: getString(job.TemplateData, "inviter_email"),
			GroupName:    getString(job.TemplateData, "group_name"),
			InviteURL:    getString(job.TemplateData, "invite_url"),
			ExpiresIn:    getString(job.TemplateData, "expires_in"),
		}
	default:
		return "", "", domainerror.NewEmailError(
			domainerror.ErrCodeInvalidTemplate,
			"unknown template type",
			domainerror.ErrInvalidTemplate,
		)
	}

	return w.renderer.Render(templateName, data)
}

// handleFailure handles a failed email job.
func (w *Worker) handleFailure(ctx context.Context, job *entity.EmailJob, err error, permanent bool) {
	job.MarkFailed(err, permanent)

	if updateErr := w.queue.Update(ctx, job); updateErr != nil {
		slog.Error("Failed to update job after failure",
			"job_id", job.ID,
			"error", updateErr,
		)
	}

	if job.Status == entity.EmailStatusFailed {
		slog.Warn("Email job permanently failed",
			"job_id", job.ID,
			"attempts", job.Attempts,
			"last_error", job.LastError,
		)
	} else {
		slog.Info("Email job scheduled for retry",
			"job_id", job.ID,
			"attempts", job.Attempts,
			"scheduled_at", job.ScheduledAt,
		)
	}
}

// getString safely extracts a string from a map.
func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ProcessNow processes all pending emails immediately (useful for testing).
func (w *Worker) ProcessNow(ctx context.Context) {
	w.processBatch(ctx)
}

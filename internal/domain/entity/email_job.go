// Package entity defines the core business entities for the domain layer.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// EmailStatus represents the status of an email job in the queue.
type EmailStatus string

const (
	EmailStatusPending    EmailStatus = "pending"
	EmailStatusProcessing EmailStatus = "processing"
	EmailStatusSent       EmailStatus = "sent"
	EmailStatusFailed     EmailStatus = "failed"
)

// EmailTemplateType represents the type of email template.
type EmailTemplateType string

const (
	TemplatePasswordReset   EmailTemplateType = "password_reset"
	TemplateGroupInvitation EmailTemplateType = "group_invitation"
)

// EmailJob represents an email in the queue waiting to be sent.
type EmailJob struct {
	ID             uuid.UUID
	TemplateType   EmailTemplateType
	RecipientEmail string
	RecipientName  string
	Subject        string
	TemplateData   map[string]interface{}
	Status         EmailStatus
	Attempts       int
	MaxAttempts    int
	LastError      string
	ResendID       string
	CreatedAt      time.Time
	ScheduledAt    time.Time
	ProcessedAt    *time.Time
}

// NewEmailJob creates a new EmailJob with default values.
func NewEmailJob(templateType EmailTemplateType, recipientEmail, recipientName, subject string, data map[string]interface{}) *EmailJob {
	now := time.Now().UTC()
	return &EmailJob{
		ID:             uuid.New(),
		TemplateType:   templateType,
		RecipientEmail: recipientEmail,
		RecipientName:  recipientName,
		Subject:        subject,
		TemplateData:   data,
		Status:         EmailStatusPending,
		Attempts:       0,
		MaxAttempts:    3,
		CreatedAt:      now,
		ScheduledAt:    now,
	}
}

// MarkProcessing marks the email job as currently being processed.
func (e *EmailJob) MarkProcessing() {
	e.Status = EmailStatusProcessing
}

// MarkSent marks the email job as successfully sent.
func (e *EmailJob) MarkSent(resendID string) {
	e.Status = EmailStatusSent
	e.ResendID = resendID
	now := time.Now().UTC()
	e.ProcessedAt = &now
}

// MarkFailed marks the email job as failed and schedules a retry if attempts remain.
func (e *EmailJob) MarkFailed(err error, permanent bool) {
	e.Attempts++
	e.LastError = err.Error()

	if permanent || e.Attempts >= e.MaxAttempts {
		e.Status = EmailStatusFailed
		now := time.Now().UTC()
		e.ProcessedAt = &now
	} else {
		e.Status = EmailStatusPending
		e.ScheduledAt = e.calculateNextRetry()
	}
}

// calculateNextRetry calculates the next retry time using exponential backoff.
// Retry delays: 0s (immediate), 1min, 5min
func (e *EmailJob) calculateNextRetry() time.Time {
	delays := []time.Duration{0, 1 * time.Minute, 5 * time.Minute}
	if e.Attempts < len(delays) {
		return time.Now().UTC().Add(delays[e.Attempts])
	}
	return time.Now().UTC().Add(5 * time.Minute)
}

// CanRetry returns true if the email job can be retried.
func (e *EmailJob) CanRetry() bool {
	return e.Attempts < e.MaxAttempts
}

// IsReadyToProcess returns true if the email job is ready to be processed.
func (e *EmailJob) IsReadyToProcess() bool {
	return e.Status == EmailStatusPending && time.Now().UTC().After(e.ScheduledAt)
}

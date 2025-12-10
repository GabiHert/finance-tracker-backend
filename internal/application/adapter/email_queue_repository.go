// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// EmailQueueRepository defines the interface for email queue persistence operations.
type EmailQueueRepository interface {
	// Create adds a new email job to the queue.
	Create(ctx context.Context, job *entity.EmailJob) error

	// GetPendingJobs retrieves jobs ready to be processed, ordered by scheduled_at.
	GetPendingJobs(ctx context.Context, limit int) ([]*entity.EmailJob, error)

	// Update saves changes to an email job.
	Update(ctx context.Context, job *entity.EmailJob) error

	// GetByID retrieves a specific job by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.EmailJob, error)

	// GetByRecipient retrieves jobs for a specific email address (for testing/debugging).
	GetByRecipient(ctx context.Context, email string) ([]*entity.EmailJob, error)

	// DeleteOldSentJobs removes sent jobs older than the specified duration (for cleanup).
	DeleteOldSentJobs(ctx context.Context, olderThanDays int) (int64, error)
}

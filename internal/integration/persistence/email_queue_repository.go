// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

// emailQueueRepository implements the adapter.EmailQueueRepository interface.
type emailQueueRepository struct {
	db *gorm.DB
}

// NewEmailQueueRepository creates a new email queue repository instance.
func NewEmailQueueRepository(db *gorm.DB) adapter.EmailQueueRepository {
	return &emailQueueRepository{
		db: db,
	}
}

// Create adds a new email job to the queue.
func (r *emailQueueRepository) Create(ctx context.Context, job *entity.EmailJob) error {
	emailModel := model.EmailQueueModelFromEntity(job)
	result := r.db.WithContext(ctx).Create(emailModel)
	if result.Error != nil {
		return domainerror.NewEmailError(
			domainerror.ErrCodeEmailQueueFailed,
			"failed to create email job",
			result.Error,
		)
	}
	return nil
}

// GetPendingJobs retrieves jobs ready to be processed.
func (r *emailQueueRepository) GetPendingJobs(ctx context.Context, limit int) ([]*entity.EmailJob, error) {
	var models []model.EmailQueueModel

	result := r.db.WithContext(ctx).
		Where("status = ?", entity.EmailStatusPending).
		Where("scheduled_at <= ?", time.Now().UTC()).
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&models)

	if result.Error != nil {
		return nil, result.Error
	}

	jobs := make([]*entity.EmailJob, len(models))
	for i, m := range models {
		jobs[i] = m.ToEntity()
	}

	return jobs, nil
}

// Update saves changes to an email job.
func (r *emailQueueRepository) Update(ctx context.Context, job *entity.EmailJob) error {
	emailModel := model.EmailQueueModelFromEntity(job)
	result := r.db.WithContext(ctx).Save(emailModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID retrieves a specific job by its ID.
func (r *emailQueueRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.EmailJob, error) {
	var emailModel model.EmailQueueModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&emailModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrEmailJobNotFound
		}
		return nil, result.Error
	}
	return emailModel.ToEntity(), nil
}

// GetByRecipient retrieves jobs for a specific email address.
func (r *emailQueueRepository) GetByRecipient(ctx context.Context, email string) ([]*entity.EmailJob, error) {
	var models []model.EmailQueueModel
	result := r.db.WithContext(ctx).
		Where("recipient_email = ?", email).
		Order("created_at DESC").
		Find(&models)

	if result.Error != nil {
		return nil, result.Error
	}

	jobs := make([]*entity.EmailJob, len(models))
	for i, m := range models {
		jobs[i] = m.ToEntity()
	}

	return jobs, nil
}

// DeleteOldSentJobs removes sent jobs older than the specified duration.
func (r *emailQueueRepository) DeleteOldSentJobs(ctx context.Context, olderThanDays int) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -olderThanDays)

	result := r.db.WithContext(ctx).
		Where("status = ?", entity.EmailStatusSent).
		Where("processed_at < ?", cutoff).
		Delete(&model.EmailQueueModel{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

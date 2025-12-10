// Package model defines database models for persistence layer.
package model

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// EmailQueueModel represents the email_queue table in the database.
type EmailQueueModel struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey"`
	TemplateType   string       `gorm:"type:varchar(50);not null"`
	RecipientEmail string       `gorm:"type:varchar(255);not null"`
	RecipientName  string       `gorm:"type:varchar(255)"`
	Subject        string       `gorm:"type:varchar(500);not null"`
	TemplateData   string       `gorm:"type:jsonb;not null;default:'{}'"`
	Status         string       `gorm:"type:varchar(20);not null;default:'pending'"`
	Attempts       int          `gorm:"not null;default:0"`
	MaxAttempts    int          `gorm:"not null;default:3"`
	LastError      string       `gorm:"type:text"`
	ResendID       string       `gorm:"type:varchar(100)"`
	CreatedAt      time.Time    `gorm:"not null"`
	ScheduledAt    time.Time    `gorm:"not null"`
	ProcessedAt    sql.NullTime `gorm:"type:timestamptz"`
}

// TableName returns the table name for the EmailQueueModel.
func (EmailQueueModel) TableName() string {
	return "email_queue"
}

// ToEntity converts an EmailQueueModel to a domain EmailJob entity.
func (m *EmailQueueModel) ToEntity() *entity.EmailJob {
	// Parse template data from JSON
	var templateData map[string]interface{}
	if m.TemplateData != "" {
		if err := json.Unmarshal([]byte(m.TemplateData), &templateData); err != nil {
			slog.Warn("Failed to unmarshal email template data", "error", err, "id", m.ID)
		}
	}
	if templateData == nil {
		templateData = make(map[string]interface{})
	}

	// Convert sql.NullTime to *time.Time
	var processedAt *time.Time
	if m.ProcessedAt.Valid {
		processedAt = &m.ProcessedAt.Time
	}

	return &entity.EmailJob{
		ID:             m.ID,
		TemplateType:   entity.EmailTemplateType(m.TemplateType),
		RecipientEmail: m.RecipientEmail,
		RecipientName:  m.RecipientName,
		Subject:        m.Subject,
		TemplateData:   templateData,
		Status:         entity.EmailStatus(m.Status),
		Attempts:       m.Attempts,
		MaxAttempts:    m.MaxAttempts,
		LastError:      m.LastError,
		ResendID:       m.ResendID,
		CreatedAt:      m.CreatedAt,
		ScheduledAt:    m.ScheduledAt,
		ProcessedAt:    processedAt,
	}
}

// EmailQueueModelFromEntity creates an EmailQueueModel from a domain EmailJob entity.
func EmailQueueModelFromEntity(job *entity.EmailJob) *EmailQueueModel {
	// Serialize template data to JSON - fallback to empty object on error
	templateDataJSON, err := json.Marshal(job.TemplateData)
	if err != nil {
		slog.Error("Failed to marshal email template data", "error", err, "job_id", job.ID)
		templateDataJSON = []byte("{}")
	}

	// Convert *time.Time to sql.NullTime
	var processedAt sql.NullTime
	if job.ProcessedAt != nil {
		processedAt = sql.NullTime{Time: *job.ProcessedAt, Valid: true}
	}

	return &EmailQueueModel{
		ID:             job.ID,
		TemplateType:   string(job.TemplateType),
		RecipientEmail: job.RecipientEmail,
		RecipientName:  job.RecipientName,
		Subject:        job.Subject,
		TemplateData:   string(templateDataJSON),
		Status:         string(job.Status),
		Attempts:       job.Attempts,
		MaxAttempts:    job.MaxAttempts,
		LastError:      job.LastError,
		ResendID:       job.ResendID,
		CreatedAt:      job.CreatedAt,
		ScheduledAt:    job.ScheduledAt,
		ProcessedAt:    processedAt,
	}
}

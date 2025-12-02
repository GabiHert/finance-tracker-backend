// Package model defines database models for persistence layer.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// GoalModel represents the goals table in the database.
type GoalModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	CategoryID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	LimitAmount   float64        `gorm:"type:decimal(15,2);not null"`
	AlertOnExceed bool           `gorm:"not null;default:true"`
	Period        string         `gorm:"type:varchar(20);not null;default:'monthly'"`
	StartDate     *time.Time     `gorm:"type:date"`
	EndDate       *time.Time     `gorm:"type:date"`
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
	DeletedAt     gorm.DeletedAt `gorm:"index"` // Soft-delete support
}

// TableName returns the table name for the GoalModel.
func (GoalModel) TableName() string {
	return "goals"
}

// ToEntity converts a GoalModel to a domain Goal entity.
func (m *GoalModel) ToEntity() *entity.Goal {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &entity.Goal{
		ID:            m.ID,
		UserID:        m.UserID,
		CategoryID:    m.CategoryID,
		LimitAmount:   m.LimitAmount,
		AlertOnExceed: m.AlertOnExceed,
		Period:        entity.GoalPeriod(m.Period),
		StartDate:     m.StartDate,
		EndDate:       m.EndDate,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}

// GoalFromEntity creates a GoalModel from a domain Goal entity.
func GoalFromEntity(goal *entity.Goal) *GoalModel {
	var deletedAt gorm.DeletedAt
	if goal.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *goal.DeletedAt, Valid: true}
	}

	return &GoalModel{
		ID:            goal.ID,
		UserID:        goal.UserID,
		CategoryID:    goal.CategoryID,
		LimitAmount:   goal.LimitAmount,
		AlertOnExceed: goal.AlertOnExceed,
		Period:        string(goal.Period),
		StartDate:     goal.StartDate,
		EndDate:       goal.EndDate,
		CreatedAt:     goal.CreatedAt,
		UpdatedAt:     goal.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}

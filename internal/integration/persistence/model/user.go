// Package model defines database models for persistence layer.
package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// UserModel represents the user table in the database.
type UserModel struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email              string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Name               string    `gorm:"type:varchar(100);not null"`
	PasswordHash       string    `gorm:"type:varchar(255);not null"`
	DateFormat         string    `gorm:"type:varchar(20);default:'YYYY-MM-DD'"`
	NumberFormat       string    `gorm:"type:varchar(10);default:'US'"`
	FirstDayOfWeek     string    `gorm:"type:varchar(10);default:'sunday'"`
	EmailNotifications bool      `gorm:"default:true"`
	GoalAlerts         bool      `gorm:"default:true"`
	RecurringReminders bool      `gorm:"default:true"`
	TermsAcceptedAt    time.Time `gorm:"not null"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
}

// TableName returns the table name for the UserModel.
func (UserModel) TableName() string {
	return "users"
}

// ToEntity converts a UserModel to a domain User entity.
func (m *UserModel) ToEntity() *entity.User {
	return &entity.User{
		ID:                 m.ID,
		Email:              m.Email,
		Name:               m.Name,
		PasswordHash:       m.PasswordHash,
		DateFormat:         entity.DateFormat(m.DateFormat),
		NumberFormat:       entity.NumberFormat(m.NumberFormat),
		FirstDayOfWeek:     entity.FirstDayOfWeek(m.FirstDayOfWeek),
		EmailNotifications: m.EmailNotifications,
		GoalAlerts:         m.GoalAlerts,
		RecurringReminders: m.RecurringReminders,
		TermsAcceptedAt:    m.TermsAcceptedAt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

// FromEntity creates a UserModel from a domain User entity.
func FromEntity(user *entity.User) *UserModel {
	return &UserModel{
		ID:                 user.ID,
		Email:              user.Email,
		Name:               user.Name,
		PasswordHash:       user.PasswordHash,
		DateFormat:         string(user.DateFormat),
		NumberFormat:       string(user.NumberFormat),
		FirstDayOfWeek:     string(user.FirstDayOfWeek),
		EmailNotifications: user.EmailNotifications,
		GoalAlerts:         user.GoalAlerts,
		RecurringReminders: user.RecurringReminders,
		TermsAcceptedAt:    user.TermsAcceptedAt,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
}

// RefreshTokenModel represents the refresh_tokens table for token invalidation tracking.
type RefreshTokenModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Token       string    `gorm:"type:varchar(500);uniqueIndex;not null"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null"`
	Invalidated bool      `gorm:"default:false"`
	ExpiresAt   time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
}

// TableName returns the table name for the RefreshTokenModel.
func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// PasswordResetTokenModel represents the password_reset_tokens table.
type PasswordResetTokenModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Token     string     `gorm:"type:varchar(500);uniqueIndex;not null"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null"`
	Email     string     `gorm:"type:varchar(255);not null"`
	Used      bool       `gorm:"default:false"`
	UsedAt    *time.Time `gorm:"type:timestamptz"`
	ExpiresAt time.Time  `gorm:"not null"`
	CreatedAt time.Time  `gorm:"not null"`
}

// TableName returns the table name for the PasswordResetTokenModel.
func (PasswordResetTokenModel) TableName() string {
	return "password_reset_tokens"
}

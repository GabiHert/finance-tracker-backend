// Package entity defines the core business entities for the domain layer.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// DateFormat represents the user's preferred date format.
type DateFormat string

const (
	DateFormatDMY DateFormat = "DD/MM/YYYY"
	DateFormatMDY DateFormat = "MM/DD/YYYY"
	DateFormatYMD DateFormat = "YYYY-MM-DD"
)

// NumberFormat represents the user's preferred number format.
type NumberFormat string

const (
	NumberFormatBR NumberFormat = "BR" // 1.234,56
	NumberFormatUS NumberFormat = "US" // 1,234.56
)

// FirstDayOfWeek represents the user's preferred first day of the week.
type FirstDayOfWeek string

const (
	FirstDayOfWeekSunday FirstDayOfWeek = "sunday"
	FirstDayOfWeekMonday FirstDayOfWeek = "monday"
)

// User represents a user in the Finance Tracker system.
type User struct {
	ID                 uuid.UUID
	Email              string
	Name               string
	PasswordHash       string
	DateFormat         DateFormat
	NumberFormat       NumberFormat
	FirstDayOfWeek     FirstDayOfWeek
	EmailNotifications bool
	GoalAlerts         bool
	RecurringReminders bool
	TermsAcceptedAt    time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewUser creates a new User with default values.
func NewUser(email, name, passwordHash string, termsAcceptedAt time.Time) *User {
	now := time.Now().UTC()
	return &User{
		ID:                 uuid.New(),
		Email:              email,
		Name:               name,
		PasswordHash:       passwordHash,
		DateFormat:         DateFormatYMD,
		NumberFormat:       NumberFormatUS,
		FirstDayOfWeek:     FirstDayOfWeekSunday,
		EmailNotifications: true,
		GoalAlerts:         true,
		RecurringReminders: true,
		TermsAcceptedAt:    termsAcceptedAt,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

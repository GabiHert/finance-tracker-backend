// Package entity defines the core business entities for the domain layer.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// GoalPeriod represents the period type for a spending goal.
type GoalPeriod string

const (
	GoalPeriodMonthly GoalPeriod = "monthly"
	GoalPeriodWeekly  GoalPeriod = "weekly"
	GoalPeriodYearly  GoalPeriod = "yearly"
)

// Goal represents a spending limit goal in the Finance Tracker system.
type Goal struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	CategoryID    uuid.UUID
	LimitAmount   float64
	AlertOnExceed bool
	Period        GoalPeriod
	StartDate     *time.Time
	EndDate       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time // Soft-delete support
}

// NewGoal creates a new Goal entity.
func NewGoal(userID, categoryID uuid.UUID, limitAmount float64, alertOnExceed bool, period GoalPeriod) *Goal {
	now := time.Now().UTC()

	return &Goal{
		ID:            uuid.New(),
		UserID:        userID,
		CategoryID:    categoryID,
		LimitAmount:   limitAmount,
		AlertOnExceed: alertOnExceed,
		Period:        period,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// GoalWithCategory represents a goal with its associated category.
type GoalWithCategory struct {
	Goal          *Goal
	Category      *Category
	CurrentAmount float64 // Current spending within the period
}

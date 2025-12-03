// Package entity defines the core business entities for the domain layer.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// CategoryRule represents an auto-categorization rule in the Finance Tracker system.
// Rules are applied to transaction descriptions using regex patterns to automatically
// assign categories to new or imported transactions.
type CategoryRule struct {
	ID         uuid.UUID
	Pattern    string     // Regex pattern to match against transaction descriptions
	CategoryID uuid.UUID  // The category to assign when the pattern matches
	Priority   int        // Higher priority rules are checked first
	IsActive   bool       // Allows disabling rules without deleting them
	OwnerType  OwnerType  // 'user' or 'group'
	OwnerID    uuid.UUID  // The owning user or group ID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time // Soft-delete support
}

// NewCategoryRule creates a new CategoryRule entity.
func NewCategoryRule(
	pattern string,
	categoryID uuid.UUID,
	priority int,
	ownerType OwnerType,
	ownerID uuid.UUID,
) *CategoryRule {
	now := time.Now().UTC()

	return &CategoryRule{
		ID:         uuid.New(),
		Pattern:    pattern,
		CategoryID: categoryID,
		Priority:   priority,
		IsActive:   true, // New rules are active by default
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// CategoryRuleWithCategory represents a category rule with its associated category.
type CategoryRuleWithCategory struct {
	Rule     *CategoryRule
	Category *Category
}

// RulePriorityUpdate represents a priority update for a single rule.
type RulePriorityUpdate struct {
	ID       uuid.UUID
	Priority int
}

// MatchingTransaction represents a transaction that matches a pattern during testing.
type MatchingTransaction struct {
	ID          uuid.UUID
	Description string
	Amount      string
	Date        time.Time
}

// PatternTestResult represents the result of testing a pattern against transactions.
type PatternTestResult struct {
	MatchingTransactions []*MatchingTransaction
	MatchCount           int
}

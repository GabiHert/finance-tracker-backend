// Package entity defines the core business entities for the domain layer.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// SuggestionStatus represents the status of an AI categorization suggestion.
type SuggestionStatus string

const (
	SuggestionStatusPending  SuggestionStatus = "pending"
	SuggestionStatusApproved SuggestionStatus = "approved"
	SuggestionStatusRejected SuggestionStatus = "rejected"
	SuggestionStatusSkipped  SuggestionStatus = "skipped"
)

// MatchType represents the type of pattern matching for AI suggestions.
type MatchType string

const (
	MatchTypeContains   MatchType = "contains"
	MatchTypeStartsWith MatchType = "startsWith"
	MatchTypeExact      MatchType = "exact"
)

// SuggestedCategoryNew represents a new category suggestion from AI.
type SuggestedCategoryNew struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// AISuggestion represents an AI-generated categorization suggestion.
// The suggestion includes a pattern match for creating a category rule,
// and can suggest either an existing category or a new one.
type AISuggestion struct {
	ID                     uuid.UUID
	UserID                 uuid.UUID
	TransactionID          uuid.UUID             // Primary transaction that triggered suggestion
	SuggestedCategoryID    *uuid.UUID            // For existing category (nullable)
	SuggestedCategoryNew   *SuggestedCategoryNew // For new category (nullable)
	MatchType              MatchType
	MatchKeyword           string
	AffectedTransactionIDs []uuid.UUID
	Status                 SuggestionStatus
	PreviousSuggestion     *string // JSON for retry context
	RetryReason            *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// NewAISuggestion creates a new AISuggestion entity with an existing category.
func NewAISuggestion(
	userID uuid.UUID,
	transactionID uuid.UUID,
	categoryID uuid.UUID,
	matchType MatchType,
	matchKeyword string,
	affectedTransactionIDs []uuid.UUID,
) *AISuggestion {
	now := time.Now().UTC()

	return &AISuggestion{
		ID:                     uuid.New(),
		UserID:                 userID,
		TransactionID:          transactionID,
		SuggestedCategoryID:    &categoryID,
		MatchType:              matchType,
		MatchKeyword:           matchKeyword,
		AffectedTransactionIDs: affectedTransactionIDs,
		Status:                 SuggestionStatusPending,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
}

// NewAISuggestionWithNewCategory creates a new AISuggestion entity with a new category.
func NewAISuggestionWithNewCategory(
	userID uuid.UUID,
	transactionID uuid.UUID,
	suggestedCategory SuggestedCategoryNew,
	matchType MatchType,
	matchKeyword string,
	affectedTransactionIDs []uuid.UUID,
) *AISuggestion {
	now := time.Now().UTC()

	return &AISuggestion{
		ID:                     uuid.New(),
		UserID:                 userID,
		TransactionID:          transactionID,
		SuggestedCategoryNew:   &suggestedCategory,
		MatchType:              matchType,
		MatchKeyword:           matchKeyword,
		AffectedTransactionIDs: affectedTransactionIDs,
		Status:                 SuggestionStatusPending,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
}

// AISuggestionWithDetails represents an AI suggestion with associated transaction and category details.
type AISuggestionWithDetails struct {
	Suggestion              *AISuggestion
	Transaction             *Transaction
	Category                *Category // Only populated if SuggestedCategoryID is set
	AffectedTransactions    []*Transaction
	AffectedTransactionCount int
}

// CategorizationStatus represents the status of the AI categorization process.
type CategorizationStatus struct {
	UncategorizedCount      int
	IsProcessing            bool
	PendingSuggestionsCount int
	JobID                   *string
}

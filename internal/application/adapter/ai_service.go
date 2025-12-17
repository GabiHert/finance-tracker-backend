// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// AICategorizationRequest represents a request to categorize transactions.
type AICategorizationRequest struct {
	UserID            uuid.UUID
	Transactions      []*TransactionForAI
	ExistingCategories []*CategoryForAI
}

// TransactionForAI represents transaction data for AI processing.
type TransactionForAI struct {
	ID          uuid.UUID
	Description string
	Amount      string
	Date        string
	Type        string
}

// CategoryForAI represents category data for AI processing.
type CategoryForAI struct {
	ID    uuid.UUID
	Name  string
	Type  string
	Icon  string
	Color string
}

// AICategorizationResult represents the AI's categorization suggestion.
type AICategorizationResult struct {
	TransactionID          uuid.UUID
	SuggestedCategoryID    *uuid.UUID             // For existing category
	SuggestedCategoryNew   *entity.SuggestedCategoryNew // For new category
	MatchType              entity.MatchType
	MatchKeyword           string
	AffectedTransactionIDs []uuid.UUID
	Confidence             float64
	Reasoning              string
}

// AICategorizationService defines the interface for AI categorization operations.
type AICategorizationService interface {
	// Categorize analyzes transactions and returns categorization suggestions.
	Categorize(ctx context.Context, request *AICategorizationRequest) ([]*AICategorizationResult, error)

	// IsAvailable checks if the AI service is available and properly configured.
	IsAvailable() bool
}

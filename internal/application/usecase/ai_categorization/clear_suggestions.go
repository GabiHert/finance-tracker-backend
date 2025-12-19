// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
)

// ClearSuggestionsInput represents the input for clearing AI suggestions.
type ClearSuggestionsInput struct {
	UserID uuid.UUID
}

// ClearSuggestionsOutput represents the output of clearing AI suggestions.
type ClearSuggestionsOutput struct {
	DeletedCount int `json:"deleted_count"`
}

// ClearSuggestionsUseCase handles clearing all pending AI suggestions.
type ClearSuggestionsUseCase struct {
	suggestionRepo adapter.AISuggestionRepository
}

// NewClearSuggestionsUseCase creates a new ClearSuggestionsUseCase instance.
func NewClearSuggestionsUseCase(
	suggestionRepo adapter.AISuggestionRepository,
) *ClearSuggestionsUseCase {
	return &ClearSuggestionsUseCase{
		suggestionRepo: suggestionRepo,
	}
}

// Execute clears all pending AI suggestions for a user.
func (uc *ClearSuggestionsUseCase) Execute(ctx context.Context, input ClearSuggestionsInput) (*ClearSuggestionsOutput, error) {
	// Delete all pending suggestions for the user
	deletedCount, err := uc.suggestionRepo.DeletePendingByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &ClearSuggestionsOutput{
		DeletedCount: deletedCount,
	}, nil
}

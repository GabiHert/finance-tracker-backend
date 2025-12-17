// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// GetSuggestionsInput represents the input for getting AI suggestions.
type GetSuggestionsInput struct {
	UserID uuid.UUID
}

// CategoryOutput represents the category suggestion in the output.
type CategoryOutput struct {
	Type          string  // "existing" or "new"
	ExistingID    *string // For existing category
	ExistingName  *string
	ExistingIcon  *string
	ExistingColor *string
	NewName       *string // For new category
	NewIcon       *string
	NewColor      *string
}

// MatchOutput represents the match rule in the output.
type MatchOutput struct {
	Type    string // "contains", "startsWith", "exact"
	Keyword string
}

// AffectedTransactionOutput represents an affected transaction in the output.
type AffectedTransactionOutput struct {
	ID          string
	Description string
	Amount      float64
	Date        string
}

// SuggestionOutput represents a single suggestion in the output.
type SuggestionOutput struct {
	ID                   string
	Category             CategoryOutput
	Match                MatchOutput
	AffectedTransactions []AffectedTransactionOutput
	AffectedCount        int
	Status               string
	CreatedAt            string
}

// SkippedTransactionOutput represents a skipped transaction in the output.
type SkippedTransactionOutput struct {
	ID          string
	Description string
	Amount      float64
	Date        string
	SkipReason  string
}

// GetSuggestionsOutput represents the output of getting AI suggestions.
type GetSuggestionsOutput struct {
	Suggestions         []SuggestionOutput
	SkippedTransactions []SkippedTransactionOutput
	TotalPending        int
	TotalSkipped        int
}

// GetSuggestionsUseCase handles retrieving AI suggestions.
type GetSuggestionsUseCase struct {
	suggestionRepo adapter.AISuggestionRepository
}

// NewGetSuggestionsUseCase creates a new GetSuggestionsUseCase instance.
func NewGetSuggestionsUseCase(
	suggestionRepo adapter.AISuggestionRepository,
) *GetSuggestionsUseCase {
	return &GetSuggestionsUseCase{
		suggestionRepo: suggestionRepo,
	}
}

// Execute retrieves all pending AI suggestions for a user.
func (uc *GetSuggestionsUseCase) Execute(ctx context.Context, input GetSuggestionsInput) (*GetSuggestionsOutput, error) {
	// Get pending suggestions with details
	suggestions, err := uc.suggestionRepo.GetPendingByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Convert to output format
	outputs := make([]SuggestionOutput, 0, len(suggestions))
	for _, s := range suggestions {
		output := uc.toSuggestionOutput(s)
		outputs = append(outputs, output)
	}

	// Note: Skipped transactions feature is not yet implemented.
	// For now, we return an empty slice to match the API contract.
	skippedOutputs := make([]SkippedTransactionOutput, 0)

	return &GetSuggestionsOutput{
		Suggestions:         outputs,
		SkippedTransactions: skippedOutputs,
		TotalPending:        len(outputs),
		TotalSkipped:        0,
	}, nil
}

// toSuggestionOutput converts a domain suggestion with details to an output.
func (uc *GetSuggestionsUseCase) toSuggestionOutput(s *entity.AISuggestionWithDetails) SuggestionOutput {
	output := SuggestionOutput{
		ID:            s.Suggestion.ID.String(),
		AffectedCount: s.AffectedTransactionCount,
		Status:        string(s.Suggestion.Status),
		CreatedAt:     s.Suggestion.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Match: MatchOutput{
			Type:    string(s.Suggestion.MatchType),
			Keyword: s.Suggestion.MatchKeyword,
		},
	}

	// Set category details
	if s.Suggestion.SuggestedCategoryID != nil && s.Category != nil {
		catID := s.Suggestion.SuggestedCategoryID.String()
		output.Category = CategoryOutput{
			Type:          "existing",
			ExistingID:    &catID,
			ExistingName:  &s.Category.Name,
			ExistingIcon:  &s.Category.Icon,
			ExistingColor: &s.Category.Color,
		}
	} else if s.Suggestion.SuggestedCategoryNew != nil {
		output.Category = CategoryOutput{
			Type:     "new",
			NewName:  &s.Suggestion.SuggestedCategoryNew.Name,
			NewIcon:  &s.Suggestion.SuggestedCategoryNew.Icon,
			NewColor: &s.Suggestion.SuggestedCategoryNew.Color,
		}
	}

	// Set affected transactions with full details
	output.AffectedTransactions = make([]AffectedTransactionOutput, 0, len(s.AffectedTransactions))
	for _, t := range s.AffectedTransactions {
		if t != nil {
			amount, _ := t.Amount.Float64()
			output.AffectedTransactions = append(output.AffectedTransactions, AffectedTransactionOutput{
				ID:          t.ID.String(),
				Description: t.Description,
				Amount:      amount,
				Date:        t.Date.Format("2006-01-02"),
			})
		}
	}

	return output
}

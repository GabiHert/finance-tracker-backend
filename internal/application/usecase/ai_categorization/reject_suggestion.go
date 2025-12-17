// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// RejectAction defines the possible actions when rejecting a suggestion.
type RejectAction string

const (
	RejectActionSkip  RejectAction = "skip"
	RejectActionRetry RejectAction = "retry"
)

// RejectSuggestionInput represents the input for rejecting an AI suggestion.
type RejectSuggestionInput struct {
	SuggestionID uuid.UUID
	UserID       uuid.UUID
	Action       RejectAction
	RetryReason  string // Optional: reason for retry to provide context to AI
}

// RejectSuggestionOutput represents the output of rejecting an AI suggestion.
type RejectSuggestionOutput struct {
	Status         string           `json:"status"`
	Message        string           `json:"message"`
	NewSuggestion  *SuggestionOutput `json:"new_suggestion,omitempty"` // Only if retry action
}

// RejectSuggestionUseCase handles rejecting an AI suggestion.
type RejectSuggestionUseCase struct {
	suggestionRepo adapter.AISuggestionRepository
	aiService      adapter.AICategorizationService
	transactionRepo adapter.TransactionRepository
	categoryRepo    adapter.CategoryRepository
}

// NewRejectSuggestionUseCase creates a new RejectSuggestionUseCase instance.
func NewRejectSuggestionUseCase(
	suggestionRepo adapter.AISuggestionRepository,
	aiService adapter.AICategorizationService,
	transactionRepo adapter.TransactionRepository,
	categoryRepo adapter.CategoryRepository,
) *RejectSuggestionUseCase {
	return &RejectSuggestionUseCase{
		suggestionRepo:  suggestionRepo,
		aiService:       aiService,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

// Execute rejects an AI suggestion by either skipping or retrying.
func (uc *RejectSuggestionUseCase) Execute(ctx context.Context, input RejectSuggestionInput) (*RejectSuggestionOutput, error) {
	// Validate action
	if input.Action != RejectActionSkip && input.Action != RejectActionRetry {
		return nil, domainerror.NewAISuggestionError(
			domainerror.ErrCodeAIInvalidAction,
			"Invalid action. Must be 'skip' or 'retry'",
			domainerror.ErrAIInvalidAction,
		)
	}

	// Get the suggestion
	suggestion, err := uc.suggestionRepo.GetByID(ctx, input.SuggestionID)
	if err != nil {
		return nil, domainerror.NewAISuggestionError(
			domainerror.ErrCodeAISuggestionNotFound,
			"AI suggestion not found",
			domainerror.ErrAISuggestionNotFound,
		)
	}

	// Verify ownership
	if suggestion.UserID != input.UserID {
		return nil, domainerror.NewAISuggestionError(
			domainerror.ErrCodeAISuggestionNotFound,
			"AI suggestion not found",
			domainerror.ErrAISuggestionNotFound,
		)
	}

	// Check if already processed
	if suggestion.Status != entity.SuggestionStatusPending {
		return nil, domainerror.NewAISuggestionError(
			domainerror.ErrCodeAISuggestionAlreadyProcessed,
			"Suggestion has already been processed",
			domainerror.ErrAISuggestionAlreadyProcessed,
		)
	}

	switch input.Action {
	case RejectActionSkip:
		return uc.handleSkip(ctx, suggestion)
	case RejectActionRetry:
		return uc.handleRetry(ctx, suggestion, input.RetryReason)
	default:
		return nil, domainerror.NewAISuggestionError(
			domainerror.ErrCodeAIInvalidAction,
			"Invalid action",
			domainerror.ErrAIInvalidAction,
		)
	}
}

// handleSkip marks the suggestion as skipped.
func (uc *RejectSuggestionUseCase) handleSkip(ctx context.Context, suggestion *entity.AISuggestion) (*RejectSuggestionOutput, error) {
	suggestion.Status = entity.SuggestionStatusSkipped
	suggestion.UpdatedAt = time.Now().UTC()

	if err := uc.suggestionRepo.Update(ctx, suggestion); err != nil {
		return nil, fmt.Errorf("failed to update suggestion status: %w", err)
	}

	return &RejectSuggestionOutput{
		Status:  string(entity.SuggestionStatusSkipped),
		Message: "Suggestion skipped successfully",
	}, nil
}

// handleRetry requests a new suggestion from the AI with additional context.
func (uc *RejectSuggestionUseCase) handleRetry(ctx context.Context, suggestion *entity.AISuggestion, retryReason string) (*RejectSuggestionOutput, error) {
	// Store the previous suggestion as JSON for context
	previousJSON, err := json.Marshal(suggestion)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize previous suggestion: %w", err)
	}
	previousStr := string(previousJSON)

	// Mark current suggestion as rejected
	suggestion.Status = entity.SuggestionStatusRejected
	suggestion.UpdatedAt = time.Now().UTC()

	if err := uc.suggestionRepo.Update(ctx, suggestion); err != nil {
		return nil, fmt.Errorf("failed to update suggestion status: %w", err)
	}

	// Get the transaction
	transaction, err := uc.transactionRepo.FindByID(ctx, suggestion.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Get existing categories
	categories, err := uc.categoryRepo.FindByOwner(ctx, entity.OwnerTypeUser, suggestion.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// Prepare request for AI with previous context
	txForAI := &adapter.TransactionForAI{
		ID:          transaction.ID,
		Description: transaction.Description,
		Amount:      transaction.Amount.String(),
		Date:        transaction.Date.Format("2006-01-02"),
		Type:        string(transaction.Type),
	}

	catsForAI := make([]*adapter.CategoryForAI, len(categories))
	for i, cat := range categories {
		catsForAI[i] = &adapter.CategoryForAI{
			ID:    cat.ID,
			Name:  cat.Name,
			Type:  string(cat.Type),
			Icon:  cat.Icon,
			Color: cat.Color,
		}
	}

	request := &adapter.AICategorizationRequest{
		UserID:             suggestion.UserID,
		Transactions:       []*adapter.TransactionForAI{txForAI},
		ExistingCategories: catsForAI,
	}

	// Call AI service for new suggestion
	results, err := uc.aiService.Categorize(ctx, request)
	if err != nil {
		return nil, domainerror.NewAISuggestionError(
			domainerror.ErrCodeAIRetryFailed,
			"Failed to get new suggestion from AI",
			domainerror.ErrAIRetryFailed,
		)
	}

	if len(results) == 0 {
		return &RejectSuggestionOutput{
			Status:  string(entity.SuggestionStatusRejected),
			Message: "No new suggestion available from AI",
		}, nil
	}

	// Create new suggestion from AI result
	result := results[0]
	var newSuggestion *entity.AISuggestion

	if result.SuggestedCategoryID != nil {
		newSuggestion = entity.NewAISuggestion(
			suggestion.UserID,
			result.TransactionID,
			*result.SuggestedCategoryID,
			result.MatchType,
			result.MatchKeyword,
			result.AffectedTransactionIDs,
		)
	} else if result.SuggestedCategoryNew != nil {
		newSuggestion = entity.NewAISuggestionWithNewCategory(
			suggestion.UserID,
			result.TransactionID,
			*result.SuggestedCategoryNew,
			result.MatchType,
			result.MatchKeyword,
			result.AffectedTransactionIDs,
		)
	}

	if newSuggestion != nil {
		// Store context about the previous suggestion
		newSuggestion.PreviousSuggestion = &previousStr
		if retryReason != "" {
			newSuggestion.RetryReason = &retryReason
		}

		if err := uc.suggestionRepo.Create(ctx, newSuggestion); err != nil {
			return nil, fmt.Errorf("failed to create new suggestion: %w", err)
		}

		// Get the new suggestion with details for output
		newSuggestionWithDetails, err := uc.suggestionRepo.GetByIDWithDetails(ctx, newSuggestion.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get new suggestion details: %w", err)
		}

		outputSuggestion := toSuggestionOutputFromDetails(newSuggestionWithDetails)

		return &RejectSuggestionOutput{
			Status:        string(entity.SuggestionStatusRejected),
			Message:       "New suggestion generated",
			NewSuggestion: &outputSuggestion,
		}, nil
	}

	return &RejectSuggestionOutput{
		Status:  string(entity.SuggestionStatusRejected),
		Message: "Suggestion rejected, no alternative available",
	}, nil
}

// toSuggestionOutputFromDetails converts suggestion with details to output format.
func toSuggestionOutputFromDetails(s *entity.AISuggestionWithDetails) SuggestionOutput {
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

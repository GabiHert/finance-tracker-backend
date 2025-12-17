// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	aicategorization "github.com/finance-tracker/backend/internal/application/usecase/ai_categorization"
)

// =============================================================================
// Request DTOs
// =============================================================================

// StartCategorizationRequest represents the request body for starting AI categorization.
type StartCategorizationRequest struct {
	// Currently empty, but can be extended with options like:
	// MaxSuggestions int `json:"max_suggestions,omitempty"`
	// IncludeNewCategories bool `json:"include_new_categories,omitempty"`
}

// ApproveSuggestionRequest represents the request body for approving a suggestion.
type ApproveSuggestionRequest struct {
	// Currently empty, auto-approves with all defaults
	// Could be extended to allow modifications before approval
}

// RejectSuggestionRequest represents the request body for rejecting a suggestion.
type RejectSuggestionRequest struct {
	Action      string `json:"action" binding:"required,oneof=skip retry"`
	RetryReason string `json:"retry_reason,omitempty"`
}

// =============================================================================
// Response DTOs
// =============================================================================

// ProcessingErrorResponse represents an AI processing error in the response.
type ProcessingErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Timestamp string `json:"timestamp"`
}

// CategorizationStatusResponse represents the response for AI categorization status.
type CategorizationStatusResponse struct {
	UncategorizedCount      int                      `json:"uncategorized_count"`
	IsProcessing            bool                     `json:"is_processing"`
	PendingSuggestionsCount int                      `json:"pending_suggestions_count"`
	JobID                   string                   `json:"job_id,omitempty"`
	HasError                bool                     `json:"has_error"`
	Error                   *ProcessingErrorResponse `json:"error,omitempty"`
}

// StartCategorizationResponse represents the response for starting AI categorization.
type StartCategorizationResponse struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

// CategorySuggestionResponse represents the category suggestion structure.
type CategorySuggestionResponse struct {
	Type          string  `json:"type"` // "existing" or "new"
	ExistingID    *string `json:"existing_id,omitempty"`
	ExistingName  *string `json:"existing_name,omitempty"`
	ExistingIcon  *string `json:"existing_icon,omitempty"`
	ExistingColor *string `json:"existing_color,omitempty"`
	NewName       *string `json:"new_name,omitempty"`
	NewIcon       *string `json:"new_icon,omitempty"`
	NewColor      *string `json:"new_color,omitempty"`
}

// MatchRuleResponse represents the match rule structure.
type MatchRuleResponse struct {
	Type    string `json:"type"` // "contains", "startsWith", "exact"
	Keyword string `json:"keyword"`
}

// AffectedTransactionResponse represents an affected transaction.
type AffectedTransactionResponse struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
}

// SuggestionResponse represents a single AI suggestion in the response.
type SuggestionResponse struct {
	ID                   string                        `json:"id"`
	Category             CategorySuggestionResponse    `json:"category"`
	Match                MatchRuleResponse             `json:"match"`
	AffectedTransactions []AffectedTransactionResponse `json:"affected_transactions"`
	AffectedCount        int                           `json:"affected_count"`
	Status               string                        `json:"status"`
	CreatedAt            string                        `json:"created_at"`
}

// SkippedTransactionResponse represents a skipped transaction.
type SkippedTransactionResponse struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	SkipReason  string  `json:"skip_reason"`
}

// SuggestionsListResponse represents the response for listing AI suggestions.
type SuggestionsListResponse struct {
	Suggestions         []SuggestionResponse         `json:"suggestions"`
	SkippedTransactions []SkippedTransactionResponse `json:"skipped_transactions"`
	TotalPending        int                          `json:"total_pending"`
	TotalSkipped        int                          `json:"total_skipped"`
}

// ApproveSuggestionResponse represents the response for approving a suggestion.
type ApproveSuggestionResponse struct {
	CategoryID            string `json:"category_id"`
	CategoryName          string `json:"category_name"`
	CategoryRuleID        string `json:"category_rule_id,omitempty"`
	CategoryRulePattern   string `json:"category_rule_pattern,omitempty"`
	TransactionsUpdated   int    `json:"transactions_updated"`
	WasNewCategoryCreated bool   `json:"was_new_category_created"`
}

// RejectSuggestionResponse represents the response for rejecting a suggestion.
type RejectSuggestionResponse struct {
	Status        string              `json:"status"`
	Message       string              `json:"message"`
	NewSuggestion *SuggestionResponse `json:"new_suggestion,omitempty"`
}

// ClearSuggestionsResponse represents the response for clearing suggestions.
type ClearSuggestionsResponse struct {
	DeletedCount int `json:"deleted_count"`
}

// =============================================================================
// Conversion Functions
// =============================================================================

// ToCategorizationStatusResponse converts use case output to DTO.
func ToCategorizationStatusResponse(output *aicategorization.GetStatusOutput) CategorizationStatusResponse {
	response := CategorizationStatusResponse{
		UncategorizedCount:      output.UncategorizedCount,
		IsProcessing:            output.IsProcessing,
		PendingSuggestionsCount: output.PendingSuggestionsCount,
		JobID:                   output.JobID,
		HasError:                output.HasError,
	}

	if output.Error != nil {
		response.Error = &ProcessingErrorResponse{
			Code:      output.Error.Code,
			Message:   output.Error.Message,
			Retryable: output.Error.Retryable,
			Timestamp: output.Error.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return response
}

// ToStartCategorizationResponse converts use case output to DTO.
func ToStartCategorizationResponse(output *aicategorization.StartCategorizationOutput) StartCategorizationResponse {
	return StartCategorizationResponse{
		JobID:   output.JobID,
		Message: output.Message,
	}
}

// ToSuggestionResponse converts use case output to DTO.
func ToSuggestionResponse(output aicategorization.SuggestionOutput) SuggestionResponse {
	// Convert affected transactions
	affectedTransactions := make([]AffectedTransactionResponse, len(output.AffectedTransactions))
	for i, t := range output.AffectedTransactions {
		affectedTransactions[i] = AffectedTransactionResponse{
			ID:          t.ID,
			Description: t.Description,
			Amount:      t.Amount,
			Date:        t.Date,
		}
	}

	return SuggestionResponse{
		ID: output.ID,
		Category: CategorySuggestionResponse{
			Type:          output.Category.Type,
			ExistingID:    output.Category.ExistingID,
			ExistingName:  output.Category.ExistingName,
			ExistingIcon:  output.Category.ExistingIcon,
			ExistingColor: output.Category.ExistingColor,
			NewName:       output.Category.NewName,
			NewIcon:       output.Category.NewIcon,
			NewColor:      output.Category.NewColor,
		},
		Match: MatchRuleResponse{
			Type:    output.Match.Type,
			Keyword: output.Match.Keyword,
		},
		AffectedTransactions: affectedTransactions,
		AffectedCount:        output.AffectedCount,
		Status:               output.Status,
		CreatedAt:            output.CreatedAt,
	}
}

// ToSkippedTransactionResponse converts use case output to DTO.
func ToSkippedTransactionResponse(output aicategorization.SkippedTransactionOutput) SkippedTransactionResponse {
	return SkippedTransactionResponse{
		ID:          output.ID,
		Description: output.Description,
		Amount:      output.Amount,
		Date:        output.Date,
		SkipReason:  output.SkipReason,
	}
}

// ToSuggestionsListResponse converts use case output to DTO.
func ToSuggestionsListResponse(output *aicategorization.GetSuggestionsOutput) SuggestionsListResponse {
	suggestions := make([]SuggestionResponse, len(output.Suggestions))
	for i, s := range output.Suggestions {
		suggestions[i] = ToSuggestionResponse(s)
	}

	skippedTransactions := make([]SkippedTransactionResponse, len(output.SkippedTransactions))
	for i, t := range output.SkippedTransactions {
		skippedTransactions[i] = ToSkippedTransactionResponse(t)
	}

	return SuggestionsListResponse{
		Suggestions:         suggestions,
		SkippedTransactions: skippedTransactions,
		TotalPending:        output.TotalPending,
		TotalSkipped:        output.TotalSkipped,
	}
}

// ToApproveSuggestionResponse converts use case output to DTO.
func ToApproveSuggestionResponse(output *aicategorization.ApproveSuggestionOutput) ApproveSuggestionResponse {
	return ApproveSuggestionResponse{
		CategoryID:            output.CategoryID,
		CategoryName:          output.CategoryName,
		CategoryRuleID:        output.CategoryRuleID,
		CategoryRulePattern:   output.CategoryRulePattern,
		TransactionsUpdated:   output.TransactionsUpdated,
		WasNewCategoryCreated: output.WasNewCategoryCreated,
	}
}

// ToRejectSuggestionResponse converts use case output to DTO.
func ToRejectSuggestionResponse(output *aicategorization.RejectSuggestionOutput) RejectSuggestionResponse {
	response := RejectSuggestionResponse{
		Status:  output.Status,
		Message: output.Message,
	}

	if output.NewSuggestion != nil {
		newSuggestion := ToSuggestionResponse(*output.NewSuggestion)
		response.NewSuggestion = &newSuggestion
	}

	return response
}

// ToClearSuggestionsResponse converts use case output to DTO.
func ToClearSuggestionsResponse(output *aicategorization.ClearSuggestionsOutput) ClearSuggestionsResponse {
	return ClearSuggestionsResponse{
		DeletedCount: output.DeletedCount,
	}
}

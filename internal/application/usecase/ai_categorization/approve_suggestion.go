// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// ApproveSuggestionInput represents the input for approving an AI suggestion.
type ApproveSuggestionInput struct {
	SuggestionID uuid.UUID
	UserID       uuid.UUID
}

// ApproveSuggestionOutput represents the output of approving an AI suggestion.
type ApproveSuggestionOutput struct {
	CategoryID              string `json:"category_id"`
	CategoryName            string `json:"category_name"`
	CategoryRuleID          string `json:"category_rule_id,omitempty"`
	CategoryRulePattern     string `json:"category_rule_pattern,omitempty"`
	TransactionsUpdated     int    `json:"transactions_updated"`
	WasNewCategoryCreated   bool   `json:"was_new_category_created"`
}

// ApproveSuggestionUseCase handles approving an AI suggestion.
type ApproveSuggestionUseCase struct {
	suggestionRepo  adapter.AISuggestionRepository
	categoryRepo    adapter.CategoryRepository
	transactionRepo adapter.TransactionRepository
	ruleRepo        adapter.CategoryRuleRepository
}

// NewApproveSuggestionUseCase creates a new ApproveSuggestionUseCase instance.
func NewApproveSuggestionUseCase(
	suggestionRepo adapter.AISuggestionRepository,
	categoryRepo adapter.CategoryRepository,
	transactionRepo adapter.TransactionRepository,
	ruleRepo adapter.CategoryRuleRepository,
) *ApproveSuggestionUseCase {
	return &ApproveSuggestionUseCase{
		suggestionRepo:  suggestionRepo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
		ruleRepo:        ruleRepo,
	}
}

// Execute approves an AI suggestion, creates category if needed, creates rule, and categorizes transactions.
func (uc *ApproveSuggestionUseCase) Execute(ctx context.Context, input ApproveSuggestionInput) (*ApproveSuggestionOutput, error) {
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

	var categoryID uuid.UUID
	var categoryName string
	wasNewCategoryCreated := false

	// Create category if it's a new category suggestion
	if suggestion.SuggestedCategoryNew != nil {
		newCategory := entity.NewCategory(
			suggestion.SuggestedCategoryNew.Name,
			suggestion.SuggestedCategoryNew.Color,
			suggestion.SuggestedCategoryNew.Icon,
			entity.OwnerTypeUser,
			input.UserID,
			entity.CategoryTypeExpense, // Default to expense, could be enhanced
		)

		if err := uc.categoryRepo.Create(ctx, newCategory); err != nil {
			return nil, fmt.Errorf("failed to create category: %w", err)
		}

		categoryID = newCategory.ID
		categoryName = newCategory.Name
		wasNewCategoryCreated = true
	} else if suggestion.SuggestedCategoryID != nil {
		categoryID = *suggestion.SuggestedCategoryID

		// Get category name
		category, err := uc.categoryRepo.FindByID(ctx, categoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get category: %w", err)
		}
		categoryName = category.Name
	} else {
		return nil, fmt.Errorf("suggestion has neither existing nor new category")
	}

	// Create the category rule based on match type and keyword
	pattern := uc.buildPattern(suggestion.MatchType, suggestion.MatchKeyword)

	rule := entity.NewCategoryRule(
		pattern,
		categoryID,
		0, // Priority will be auto-assigned
		entity.OwnerTypeUser,
		input.UserID,
	)

	// Check if pattern already exists
	exists, err := uc.ruleRepo.ExistsByPatternAndOwner(ctx, pattern, entity.OwnerTypeUser, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check pattern existence: %w", err)
	}

	var ruleID string
	var rulePattern string

	if !exists {
		// Get max priority and assign new priority
		maxPriority, err := uc.ruleRepo.GetMaxPriorityByOwner(ctx, entity.OwnerTypeUser, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get max priority: %w", err)
		}
		rule.Priority = maxPriority + 1

		if err := uc.ruleRepo.Create(ctx, rule); err != nil {
			return nil, fmt.Errorf("failed to create category rule: %w", err)
		}

		ruleID = rule.ID.String()
		rulePattern = rule.Pattern
	}

	// Categorize affected transactions
	transactionIDs := append([]uuid.UUID{suggestion.TransactionID}, suggestion.AffectedTransactionIDs...)
	updatedCount, err := uc.transactionRepo.BulkUpdateCategory(ctx, transactionIDs, categoryID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transactions: %w", err)
	}

	// Update suggestion status to approved
	suggestion.Status = entity.SuggestionStatusApproved
	suggestion.UpdatedAt = time.Now().UTC()
	if err := uc.suggestionRepo.Update(ctx, suggestion); err != nil {
		return nil, fmt.Errorf("failed to update suggestion status: %w", err)
	}

	return &ApproveSuggestionOutput{
		CategoryID:            categoryID.String(),
		CategoryName:          categoryName,
		CategoryRuleID:        ruleID,
		CategoryRulePattern:   rulePattern,
		TransactionsUpdated:   int(updatedCount),
		WasNewCategoryCreated: wasNewCategoryCreated,
	}, nil
}

// buildPattern builds a regex pattern from match type and keyword.
func (uc *ApproveSuggestionUseCase) buildPattern(matchType entity.MatchType, keyword string) string {
	switch matchType {
	case entity.MatchTypeExact:
		return fmt.Sprintf("^%s$", keyword)
	case entity.MatchTypeStartsWith:
		return fmt.Sprintf("^%s", keyword)
	case entity.MatchTypeContains:
		fallthrough
	default:
		return fmt.Sprintf("(?i)%s", keyword) // Case-insensitive contains
	}
}

// Package categoryrule contains category rule-related use cases.
package categoryrule

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

const (
	// MaxPatternLength is the maximum allowed length for regex patterns.
	MaxPatternLength = 255
)

// CreateCategoryRuleInput represents the input for category rule creation.
type CreateCategoryRuleInput struct {
	Pattern    string
	CategoryID uuid.UUID
	Priority   *int // Optional, defaults to max priority + 1
	OwnerType  entity.OwnerType
	OwnerID    uuid.UUID
}

// CreateCategoryRuleOutput represents the output of category rule creation.
type CreateCategoryRuleOutput struct {
	Rule                *entity.CategoryRuleWithCategory
	TransactionsUpdated int
}

// CreateCategoryRuleUseCase handles category rule creation logic.
type CreateCategoryRuleUseCase struct {
	ruleRepo        adapter.CategoryRuleRepository
	categoryRepo    adapter.CategoryRepository
	transactionRepo adapter.TransactionRepository
}

// NewCreateCategoryRuleUseCase creates a new CreateCategoryRuleUseCase instance.
func NewCreateCategoryRuleUseCase(
	ruleRepo adapter.CategoryRuleRepository,
	categoryRepo adapter.CategoryRepository,
	transactionRepo adapter.TransactionRepository,
) *CreateCategoryRuleUseCase {
	return &CreateCategoryRuleUseCase{
		ruleRepo:        ruleRepo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
	}
}

// Execute performs the category rule creation.
func (uc *CreateCategoryRuleUseCase) Execute(ctx context.Context, input CreateCategoryRuleInput) (*CreateCategoryRuleOutput, error) {
	// Validate pattern is not empty
	if input.Pattern == "" {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeMissingRuleFields,
			"pattern is required",
			domainerror.ErrCategoryRuleMissingFields,
		)
	}

	// Validate pattern length
	if len(input.Pattern) > MaxPatternLength {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodePatternTooLong,
			fmt.Sprintf("pattern must not exceed %d characters", MaxPatternLength),
			domainerror.ErrPatternTooLong,
		)
	}

	// Validate regex pattern
	if _, err := regexp.Compile(input.Pattern); err != nil {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeInvalidPattern,
			"invalid regex pattern: "+err.Error(),
			domainerror.ErrInvalidPattern,
		)
	}

	// Validate owner type
	if !isValidOwnerType(input.OwnerType) {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeRuleOwnerTypeMismatch,
			"owner type must be 'user' or 'group'",
			nil,
		)
	}

	// Verify category exists and belongs to the owner
	category, err := uc.categoryRepo.FindByID(ctx, input.CategoryID)
	if err != nil {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeCategoryNotFoundForRule,
			"category not found",
			domainerror.ErrCategoryNotFound,
		)
	}

	// Verify category belongs to the same owner
	if category.OwnerType != input.OwnerType || category.OwnerID != input.OwnerID {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeNotAuthorizedRule,
			"category does not belong to the rule owner",
			domainerror.ErrNotAuthorizedToModifyRule,
		)
	}

	// Check if pattern already exists for this owner
	exists, err := uc.ruleRepo.ExistsByPatternAndOwner(ctx, input.Pattern, input.OwnerType, input.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check pattern existence: %w", err)
	}
	if exists {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeCategoryRulePatternExists,
			"a rule with this pattern already exists",
			domainerror.ErrCategoryRulePatternExists,
		)
	}

	// Determine priority
	priority := 0
	if input.Priority != nil {
		priority = *input.Priority
	} else {
		// Auto-assign priority: max existing priority + 1
		maxPriority, err := uc.ruleRepo.GetMaxPriorityByOwner(ctx, input.OwnerType, input.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get max priority: %w", err)
		}
		priority = maxPriority + 1
	}

	// Create rule entity
	rule := entity.NewCategoryRule(
		input.Pattern,
		input.CategoryID,
		priority,
		input.OwnerType,
		input.OwnerID,
	)

	// Save rule to database
	if err := uc.ruleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create category rule: %w", err)
	}

	// Apply rule to existing uncategorized transactions
	updatedCount := 0
	if rule.IsActive {
		count, err := uc.transactionRepo.BulkUpdateCategoryByPattern(
			ctx,
			rule.Pattern,
			rule.CategoryID,
			input.OwnerType,
			input.OwnerID,
		)
		if err != nil {
			// Log warning but don't fail - rule was created successfully
			// Just return 0 for updated count
		} else {
			updatedCount = count
		}
	}

	return &CreateCategoryRuleOutput{
		Rule: &entity.CategoryRuleWithCategory{
			Rule:     rule,
			Category: category,
		},
		TransactionsUpdated: updatedCount,
	}, nil
}

// isValidOwnerType validates the owner type.
func isValidOwnerType(ownerType entity.OwnerType) bool {
	return ownerType == entity.OwnerTypeUser || ownerType == entity.OwnerTypeGroup
}

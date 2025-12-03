// Package categoryrule contains category rule-related use cases.
package categoryrule

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// UpdateCategoryRuleInput represents the input for category rule update.
type UpdateCategoryRuleInput struct {
	RuleID     uuid.UUID
	Pattern    *string    // Optional
	CategoryID *uuid.UUID // Optional
	Priority   *int       // Optional
	IsActive   *bool      // Optional
	OwnerType  entity.OwnerType
	OwnerID    uuid.UUID
}

// UpdateCategoryRuleOutput represents the output of category rule update.
type UpdateCategoryRuleOutput struct {
	Rule *entity.CategoryRuleWithCategory
}

// UpdateCategoryRuleUseCase handles category rule update logic.
type UpdateCategoryRuleUseCase struct {
	ruleRepo     adapter.CategoryRuleRepository
	categoryRepo adapter.CategoryRepository
}

// NewUpdateCategoryRuleUseCase creates a new UpdateCategoryRuleUseCase instance.
func NewUpdateCategoryRuleUseCase(
	ruleRepo adapter.CategoryRuleRepository,
	categoryRepo adapter.CategoryRepository,
) *UpdateCategoryRuleUseCase {
	return &UpdateCategoryRuleUseCase{
		ruleRepo:     ruleRepo,
		categoryRepo: categoryRepo,
	}
}

// Execute performs the category rule update.
func (uc *UpdateCategoryRuleUseCase) Execute(ctx context.Context, input UpdateCategoryRuleInput) (*UpdateCategoryRuleOutput, error) {
	// Find the existing rule
	rule, err := uc.ruleRepo.FindByID(ctx, input.RuleID)
	if err != nil {
		if errors.Is(err, domainerror.ErrCategoryRuleNotFound) {
			return nil, domainerror.NewCategoryRuleError(
				domainerror.ErrCodeCategoryRuleNotFound,
				"category rule not found",
				domainerror.ErrCategoryRuleNotFound,
			)
		}
		return nil, fmt.Errorf("failed to find category rule: %w", err)
	}

	// Check if user is authorized to modify this rule
	if rule.OwnerType != input.OwnerType || rule.OwnerID != input.OwnerID {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeNotAuthorizedRule,
			"not authorized to modify this rule",
			domainerror.ErrNotAuthorizedToModifyRule,
		)
	}

	// Update pattern if provided
	if input.Pattern != nil {
		// Validate pattern length
		if len(*input.Pattern) > MaxPatternLength {
			return nil, domainerror.NewCategoryRuleError(
				domainerror.ErrCodePatternTooLong,
				fmt.Sprintf("pattern must not exceed %d characters", MaxPatternLength),
				domainerror.ErrPatternTooLong,
			)
		}

		// Validate regex pattern
		if _, err := regexp.Compile(*input.Pattern); err != nil {
			return nil, domainerror.NewCategoryRuleError(
				domainerror.ErrCodeInvalidPattern,
				"invalid regex pattern: "+err.Error(),
				domainerror.ErrInvalidPattern,
			)
		}

		// Check if new pattern already exists for this owner (excluding current rule)
		if *input.Pattern != rule.Pattern {
			exists, err := uc.ruleRepo.ExistsByPatternAndOwnerExcluding(ctx, *input.Pattern, input.OwnerType, input.OwnerID, input.RuleID)
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
		}

		rule.Pattern = *input.Pattern
	}

	// Update category if provided
	var category *entity.Category
	if input.CategoryID != nil {
		// Verify category exists
		category, err = uc.categoryRepo.FindByID(ctx, *input.CategoryID)
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

		rule.CategoryID = *input.CategoryID
	} else {
		// Fetch the current category
		category, _ = uc.categoryRepo.FindByID(ctx, rule.CategoryID)
	}

	// Update priority if provided
	if input.Priority != nil {
		rule.Priority = *input.Priority
	}

	// Update is_active if provided
	if input.IsActive != nil {
		rule.IsActive = *input.IsActive
	}

	// Update timestamp
	rule.UpdatedAt = time.Now().UTC()

	// Save updated rule
	if err := uc.ruleRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update category rule: %w", err)
	}

	return &UpdateCategoryRuleOutput{
		Rule: &entity.CategoryRuleWithCategory{
			Rule:     rule,
			Category: category,
		},
	}, nil
}

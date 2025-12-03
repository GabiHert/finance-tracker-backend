// Package categoryrule contains category rule-related use cases.
package categoryrule

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// DeleteCategoryRuleInput represents the input for category rule deletion.
type DeleteCategoryRuleInput struct {
	RuleID    uuid.UUID
	OwnerType entity.OwnerType
	OwnerID   uuid.UUID
}

// DeleteCategoryRuleOutput represents the output of category rule deletion.
type DeleteCategoryRuleOutput struct {
	Success bool
}

// DeleteCategoryRuleUseCase handles category rule deletion logic.
type DeleteCategoryRuleUseCase struct {
	ruleRepo adapter.CategoryRuleRepository
}

// NewDeleteCategoryRuleUseCase creates a new DeleteCategoryRuleUseCase instance.
func NewDeleteCategoryRuleUseCase(ruleRepo adapter.CategoryRuleRepository) *DeleteCategoryRuleUseCase {
	return &DeleteCategoryRuleUseCase{
		ruleRepo: ruleRepo,
	}
}

// Execute performs the category rule deletion.
func (uc *DeleteCategoryRuleUseCase) Execute(ctx context.Context, input DeleteCategoryRuleInput) (*DeleteCategoryRuleOutput, error) {
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

	// Check if user is authorized to delete this rule
	if rule.OwnerType != input.OwnerType || rule.OwnerID != input.OwnerID {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeNotAuthorizedRule,
			"not authorized to delete this rule",
			domainerror.ErrNotAuthorizedToModifyRule,
		)
	}

	// Delete the rule
	if err := uc.ruleRepo.Delete(ctx, input.RuleID); err != nil {
		return nil, fmt.Errorf("failed to delete category rule: %w", err)
	}

	return &DeleteCategoryRuleOutput{
		Success: true,
	}, nil
}

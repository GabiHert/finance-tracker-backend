// Package categoryrule contains category rule-related use cases.
package categoryrule

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// ReorderCategoryRulesInput represents the input for reordering category rules.
type ReorderCategoryRulesInput struct {
	Order     []RulePriorityInput
	OwnerType entity.OwnerType
	OwnerID   uuid.UUID
}

// RulePriorityInput represents a priority update for a single rule.
type RulePriorityInput struct {
	ID       uuid.UUID
	Priority int
}

// ReorderCategoryRulesOutput represents the output of reordering category rules.
type ReorderCategoryRulesOutput struct {
	Rules []*CategoryRuleOutput
}

// ReorderCategoryRulesUseCase handles category rules reordering logic.
type ReorderCategoryRulesUseCase struct {
	ruleRepo adapter.CategoryRuleRepository
}

// NewReorderCategoryRulesUseCase creates a new ReorderCategoryRulesUseCase instance.
func NewReorderCategoryRulesUseCase(ruleRepo adapter.CategoryRuleRepository) *ReorderCategoryRulesUseCase {
	return &ReorderCategoryRulesUseCase{
		ruleRepo: ruleRepo,
	}
}

// Execute performs the category rules reordering.
func (uc *ReorderCategoryRulesUseCase) Execute(ctx context.Context, input ReorderCategoryRulesInput) (*ReorderCategoryRulesOutput, error) {
	// Validate that at least one rule is provided
	if len(input.Order) == 0 {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeMissingRuleFields,
			"at least one rule must be provided",
			domainerror.ErrCategoryRuleMissingFields,
		)
	}

	// Verify all rules exist and belong to the owner
	for _, update := range input.Order {
		rule, err := uc.ruleRepo.FindByID(ctx, update.ID)
		if err != nil {
			return nil, domainerror.NewCategoryRuleError(
				domainerror.ErrCodeCategoryRuleNotFound,
				fmt.Sprintf("category rule not found: %s", update.ID),
				domainerror.ErrCategoryRuleNotFound,
			)
		}

		// Check ownership
		if rule.OwnerType != input.OwnerType || rule.OwnerID != input.OwnerID {
			return nil, domainerror.NewCategoryRuleError(
				domainerror.ErrCodeNotAuthorizedRule,
				fmt.Sprintf("not authorized to modify rule: %s", update.ID),
				domainerror.ErrNotAuthorizedToModifyRule,
			)
		}
	}

	// Convert input to entity format
	updates := make([]entity.RulePriorityUpdate, len(input.Order))
	for i, update := range input.Order {
		updates[i] = entity.RulePriorityUpdate{
			ID:       update.ID,
			Priority: update.Priority,
		}
	}

	// Update priorities in batch
	if err := uc.ruleRepo.UpdatePriorities(ctx, updates); err != nil {
		return nil, fmt.Errorf("failed to update rule priorities: %w", err)
	}

	// Fetch updated rules with categories
	rulesWithCategories, err := uc.ruleRepo.FindByOwnerWithCategories(ctx, input.OwnerType, input.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated rules: %w", err)
	}

	// Build output
	output := &ReorderCategoryRulesOutput{
		Rules: make([]*CategoryRuleOutput, len(rulesWithCategories)),
	}

	for i, rwc := range rulesWithCategories {
		ruleOutput := &CategoryRuleOutput{
			ID:         rwc.Rule.ID,
			Pattern:    rwc.Rule.Pattern,
			CategoryID: rwc.Rule.CategoryID,
			Priority:   rwc.Rule.Priority,
			IsActive:   rwc.Rule.IsActive,
			OwnerType:  rwc.Rule.OwnerType,
			OwnerID:    rwc.Rule.OwnerID,
			CreatedAt:  rwc.Rule.CreatedAt,
			UpdatedAt:  rwc.Rule.UpdatedAt,
		}

		if rwc.Category != nil {
			ruleOutput.CategoryName = rwc.Category.Name
			ruleOutput.CategoryIcon = rwc.Category.Icon
			ruleOutput.CategoryColor = rwc.Category.Color
		}

		output.Rules[i] = ruleOutput
	}

	return output, nil
}

// Package categoryrule contains category rule-related use cases.
package categoryrule

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// ListCategoryRulesInput represents the input for listing category rules.
type ListCategoryRulesInput struct {
	OwnerType  entity.OwnerType
	OwnerID    uuid.UUID
	ActiveOnly bool // If true, only return active rules
}

// ListCategoryRulesOutput represents the output of listing category rules.
type ListCategoryRulesOutput struct {
	Rules []*CategoryRuleOutput
}

// CategoryRuleOutput represents a single category rule in the output.
type CategoryRuleOutput struct {
	ID           uuid.UUID
	Pattern      string
	CategoryID   uuid.UUID
	CategoryName string
	CategoryIcon string
	CategoryColor string
	Priority     int
	IsActive     bool
	OwnerType    entity.OwnerType
	OwnerID      uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ListCategoryRulesUseCase handles listing category rules logic.
type ListCategoryRulesUseCase struct {
	ruleRepo adapter.CategoryRuleRepository
}

// NewListCategoryRulesUseCase creates a new ListCategoryRulesUseCase instance.
func NewListCategoryRulesUseCase(ruleRepo adapter.CategoryRuleRepository) *ListCategoryRulesUseCase {
	return &ListCategoryRulesUseCase{
		ruleRepo: ruleRepo,
	}
}

// Execute performs the category rules listing.
func (uc *ListCategoryRulesUseCase) Execute(ctx context.Context, input ListCategoryRulesInput) (*ListCategoryRulesOutput, error) {
	// Fetch rules with their categories
	rulesWithCategories, err := uc.ruleRepo.FindByOwnerWithCategories(ctx, input.OwnerType, input.OwnerID)
	if err != nil {
		return nil, err
	}

	// Filter active only if requested
	if input.ActiveOnly {
		filtered := make([]*entity.CategoryRuleWithCategory, 0)
		for _, rwc := range rulesWithCategories {
			if rwc.Rule.IsActive {
				filtered = append(filtered, rwc)
			}
		}
		rulesWithCategories = filtered
	}

	// Build output
	output := &ListCategoryRulesOutput{
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

		// Add category info if available
		if rwc.Category != nil {
			ruleOutput.CategoryName = rwc.Category.Name
			ruleOutput.CategoryIcon = rwc.Category.Icon
			ruleOutput.CategoryColor = rwc.Category.Color
		}

		output.Rules[i] = ruleOutput
	}

	return output, nil
}

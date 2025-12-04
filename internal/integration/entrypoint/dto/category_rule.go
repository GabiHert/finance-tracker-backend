// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"time"

	categoryrule "github.com/finance-tracker/backend/internal/application/usecase/category_rule"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// CreateCategoryRuleRequest represents the request body for category rule creation.
type CreateCategoryRuleRequest struct {
	Pattern    string  `json:"pattern" binding:"required,min=1,max=255"`
	CategoryID string  `json:"category_id" binding:"required,uuid"`
	Priority   *int    `json:"priority,omitempty"`
}

// UpdateCategoryRuleRequest represents the request body for category rule update.
type UpdateCategoryRuleRequest struct {
	Pattern    *string `json:"pattern,omitempty" binding:"omitempty,min=1,max=255"`
	CategoryID *string `json:"category_id,omitempty" binding:"omitempty,uuid"`
	Priority   *int    `json:"priority,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

// ReorderCategoryRulesRequest represents the request body for reordering rules.
type ReorderCategoryRulesRequest struct {
	Order []RulePriorityItem `json:"order" binding:"required,dive"`
}

// RulePriorityItem represents a single rule priority update.
type RulePriorityItem struct {
	ID       string `json:"id" binding:"required,uuid"`
	Priority int    `json:"priority" binding:"required"`
}

// TestPatternRequest represents the request body for pattern testing.
type TestPatternRequest struct {
	Pattern string `json:"pattern" binding:"required,min=1,max=255"`
}

// CategoryRuleResponse represents a single category rule in API responses.
type CategoryRuleResponse struct {
	ID                  string    `json:"id"`
	Pattern             string    `json:"pattern"`
	CategoryID          string    `json:"category_id"`
	CategoryName        string    `json:"category_name,omitempty"`
	CategoryIcon        string    `json:"category_icon,omitempty"`
	CategoryColor       string    `json:"category_color,omitempty"`
	Priority            int       `json:"priority"`
	IsActive            bool      `json:"is_active"`
	OwnerType           string    `json:"owner_type"`
	OwnerID             string    `json:"owner_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	TransactionsUpdated int       `json:"transactions_updated,omitempty"`
}

// CategoryRuleListResponse represents the response for listing category rules.
type CategoryRuleListResponse struct {
	Rules []CategoryRuleResponse `json:"rules"`
}

// TestPatternResponse represents the response for pattern testing.
type TestPatternResponse struct {
	MatchingTransactions []MatchingTransactionResponse `json:"matching_transactions"`
	MatchCount           int                           `json:"match_count"`
}

// MatchingTransactionResponse represents a matching transaction in the response.
type MatchingTransactionResponse struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Amount      string `json:"amount"`
	Date        string `json:"date"`
}

// ToCategoryRuleResponse converts a domain CategoryRuleWithCategory to a CategoryRuleResponse DTO.
func ToCategoryRuleResponse(rwc *entity.CategoryRuleWithCategory) CategoryRuleResponse {
	response := CategoryRuleResponse{
		ID:         rwc.Rule.ID.String(),
		Pattern:    rwc.Rule.Pattern,
		CategoryID: rwc.Rule.CategoryID.String(),
		Priority:   rwc.Rule.Priority,
		IsActive:   rwc.Rule.IsActive,
		OwnerType:  string(rwc.Rule.OwnerType),
		OwnerID:    rwc.Rule.OwnerID.String(),
		CreatedAt:  rwc.Rule.CreatedAt,
		UpdatedAt:  rwc.Rule.UpdatedAt,
	}

	if rwc.Category != nil {
		response.CategoryName = rwc.Category.Name
		response.CategoryIcon = rwc.Category.Icon
		response.CategoryColor = rwc.Category.Color
	}

	return response
}

// ToCategoryRuleResponseFromOutput converts a CategoryRuleOutput to a CategoryRuleResponse DTO.
func ToCategoryRuleResponseFromOutput(output *categoryrule.CategoryRuleOutput) CategoryRuleResponse {
	return CategoryRuleResponse{
		ID:            output.ID.String(),
		Pattern:       output.Pattern,
		CategoryID:    output.CategoryID.String(),
		CategoryName:  output.CategoryName,
		CategoryIcon:  output.CategoryIcon,
		CategoryColor: output.CategoryColor,
		Priority:      output.Priority,
		IsActive:      output.IsActive,
		OwnerType:     string(output.OwnerType),
		OwnerID:       output.OwnerID.String(),
		CreatedAt:     output.CreatedAt,
		UpdatedAt:     output.UpdatedAt,
	}
}

// ToCategoryRuleListResponse converts a list of CategoryRuleOutput to CategoryRuleListResponse.
func ToCategoryRuleListResponse(outputs []*categoryrule.CategoryRuleOutput) CategoryRuleListResponse {
	rules := make([]CategoryRuleResponse, len(outputs))
	for i, output := range outputs {
		rules[i] = ToCategoryRuleResponseFromOutput(output)
	}
	return CategoryRuleListResponse{
		Rules: rules,
	}
}

// ToTestPatternResponse converts a TestPatternOutput to TestPatternResponse.
func ToTestPatternResponse(output *categoryrule.TestPatternOutput) TestPatternResponse {
	transactions := make([]MatchingTransactionResponse, len(output.MatchingTransactions))
	for i, tx := range output.MatchingTransactions {
		transactions[i] = MatchingTransactionResponse{
			ID:          tx.ID,
			Description: tx.Description,
			Amount:      tx.Amount,
			Date:        tx.Date,
		}
	}
	return TestPatternResponse{
		MatchingTransactions: transactions,
		MatchCount:           output.MatchCount,
	}
}

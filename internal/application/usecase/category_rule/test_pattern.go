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
	// DefaultMatchLimit is the default number of matching transactions to return.
	DefaultMatchLimit = 10
	// MaxMatchLimit is the maximum number of matching transactions to return.
	MaxMatchLimit = 100
)

// TestPatternInput represents the input for pattern testing.
type TestPatternInput struct {
	Pattern   string
	Limit     int // Optional, defaults to DefaultMatchLimit
	OwnerType entity.OwnerType
	OwnerID   uuid.UUID
}

// TestPatternOutput represents the output of pattern testing.
type TestPatternOutput struct {
	MatchingTransactions []*MatchingTransactionOutput
	MatchCount           int
}

// MatchingTransactionOutput represents a transaction that matches the pattern.
type MatchingTransactionOutput struct {
	ID          string
	Description string
	Amount      string
	Date        string
}

// TestPatternUseCase handles pattern testing logic.
type TestPatternUseCase struct {
	ruleRepo adapter.CategoryRuleRepository
}

// NewTestPatternUseCase creates a new TestPatternUseCase instance.
func NewTestPatternUseCase(ruleRepo adapter.CategoryRuleRepository) *TestPatternUseCase {
	return &TestPatternUseCase{
		ruleRepo: ruleRepo,
	}
}

// Execute performs the pattern testing.
func (uc *TestPatternUseCase) Execute(ctx context.Context, input TestPatternInput) (*TestPatternOutput, error) {
	// Validate pattern is not empty
	if input.Pattern == "" {
		return nil, domainerror.NewCategoryRuleError(
			domainerror.ErrCodeMissingRuleFields,
			"pattern is required",
			domainerror.ErrCategoryRuleMissingFields,
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

	// Set default limit if not provided
	limit := input.Limit
	if limit <= 0 {
		limit = DefaultMatchLimit
	} else if limit > MaxMatchLimit {
		limit = MaxMatchLimit
	}

	// Find matching transactions
	result, err := uc.ruleRepo.FindMatchingTransactions(ctx, input.Pattern, input.OwnerType, input.OwnerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching transactions: %w", err)
	}

	// Build output
	output := &TestPatternOutput{
		MatchCount:           result.MatchCount,
		MatchingTransactions: make([]*MatchingTransactionOutput, len(result.MatchingTransactions)),
	}

	for i, tx := range result.MatchingTransactions {
		output.MatchingTransactions[i] = &MatchingTransactionOutput{
			ID:          tx.ID.String(),
			Description: tx.Description,
			Amount:      tx.Amount,
			Date:        tx.Date.Format("2006-01-02"),
		}
	}

	return output, nil
}

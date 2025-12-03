// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// CategoryRuleRepository defines the interface for category rule persistence operations.
type CategoryRuleRepository interface {
	// Create creates a new category rule in the database.
	Create(ctx context.Context, rule *entity.CategoryRule) error

	// FindByID retrieves a category rule by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.CategoryRule, error)

	// FindByIDWithCategory retrieves a category rule with its category by ID.
	FindByIDWithCategory(ctx context.Context, id uuid.UUID) (*entity.CategoryRuleWithCategory, error)

	// FindByOwner retrieves all category rules for a given owner, sorted by priority (descending).
	FindByOwner(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) ([]*entity.CategoryRule, error)

	// FindByOwnerWithCategories retrieves all category rules with their categories for a given owner.
	FindByOwnerWithCategories(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) ([]*entity.CategoryRuleWithCategory, error)

	// FindActiveByOwner retrieves only active category rules for a given owner, sorted by priority (descending).
	FindActiveByOwner(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) ([]*entity.CategoryRule, error)

	// Update updates an existing category rule in the database.
	Update(ctx context.Context, rule *entity.CategoryRule) error

	// Delete removes a category rule from the database.
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByPatternAndOwner checks if a rule with the given pattern exists for the owner.
	ExistsByPatternAndOwner(ctx context.Context, pattern string, ownerType entity.OwnerType, ownerID uuid.UUID) (bool, error)

	// ExistsByPatternAndOwnerExcluding checks if a rule with the given pattern exists for the owner,
	// excluding a specific rule ID (used for updates).
	ExistsByPatternAndOwnerExcluding(ctx context.Context, pattern string, ownerType entity.OwnerType, ownerID uuid.UUID, excludeID uuid.UUID) (bool, error)

	// UpdatePriorities updates the priorities for multiple rules in a batch operation.
	UpdatePriorities(ctx context.Context, updates []entity.RulePriorityUpdate) error

	// FindMatchingTransactions finds transactions that match the given regex pattern.
	FindMatchingTransactions(ctx context.Context, pattern string, ownerType entity.OwnerType, ownerID uuid.UUID, limit int) (*entity.PatternTestResult, error)

	// GetMaxPriorityByOwner gets the maximum priority value for rules owned by the given owner.
	GetMaxPriorityByOwner(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) (int, error)
}

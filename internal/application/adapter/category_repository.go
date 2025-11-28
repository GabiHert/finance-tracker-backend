// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// CategoryRepository defines the interface for category persistence operations.
type CategoryRepository interface {
	// Create creates a new category in the database.
	Create(ctx context.Context, category *entity.Category) error

	// FindByID retrieves a category by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)

	// FindByOwner retrieves all categories for a given owner.
	FindByOwner(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID) ([]*entity.Category, error)

	// FindByOwnerAndType retrieves categories for a given owner filtered by type.
	FindByOwnerAndType(ctx context.Context, ownerType entity.OwnerType, ownerID uuid.UUID, categoryType entity.CategoryType) ([]*entity.Category, error)

	// FindByNameAndOwner retrieves a category by name and owner (for uniqueness check).
	FindByNameAndOwner(ctx context.Context, name string, ownerType entity.OwnerType, ownerID uuid.UUID) (*entity.Category, error)

	// Update updates an existing category in the database.
	Update(ctx context.Context, category *entity.Category) error

	// Delete removes a category from the database.
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByNameAndOwner checks if a category with the given name exists for the owner.
	ExistsByNameAndOwner(ctx context.Context, name string, ownerType entity.OwnerType, ownerID uuid.UUID) (bool, error)

	// GetTransactionStats retrieves transaction statistics for categories within a date range.
	GetTransactionStats(ctx context.Context, categoryIDs []uuid.UUID, startDate, endDate time.Time) (map[uuid.UUID]*CategoryStats, error)
}

// CategoryStats represents transaction statistics for a category.
type CategoryStats struct {
	TransactionCount int
	PeriodTotal      float64
}

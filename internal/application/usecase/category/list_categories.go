// Package category contains category-related use cases.
package category

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// ListCategoriesInput represents the input for listing categories.
type ListCategoriesInput struct {
	OwnerType    entity.OwnerType
	OwnerID      uuid.UUID
	CategoryType *entity.CategoryType // Optional filter by category type
	StartDate    *time.Time           // Optional start date for statistics
	EndDate      *time.Time           // Optional end date for statistics
}

// ListCategoriesOutput represents the output of listing categories.
type ListCategoriesOutput struct {
	Categories []*CategoryOutput
}

// CategoryOutput represents a single category in the output.
type CategoryOutput struct {
	ID               uuid.UUID
	Name             string
	Color            string
	Icon             string
	OwnerType        entity.OwnerType
	OwnerID          uuid.UUID
	Type             entity.CategoryType
	TransactionCount int
	PeriodTotal      float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ListCategoriesUseCase handles listing categories logic.
type ListCategoriesUseCase struct {
	categoryRepo adapter.CategoryRepository
}

// NewListCategoriesUseCase creates a new ListCategoriesUseCase instance.
func NewListCategoriesUseCase(categoryRepo adapter.CategoryRepository) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{
		categoryRepo: categoryRepo,
	}
}

// Execute performs the category listing.
func (uc *ListCategoriesUseCase) Execute(ctx context.Context, input ListCategoriesInput) (*ListCategoriesOutput, error) {
	var categories []*entity.Category
	var err error

	// Fetch categories based on filters
	if input.CategoryType != nil {
		categories, err = uc.categoryRepo.FindByOwnerAndType(ctx, input.OwnerType, input.OwnerID, *input.CategoryType)
	} else {
		categories, err = uc.categoryRepo.FindByOwner(ctx, input.OwnerType, input.OwnerID)
	}

	if err != nil {
		return nil, err
	}

	// Get transaction statistics if date range is provided
	var stats map[uuid.UUID]*adapter.CategoryStats
	if input.StartDate != nil && input.EndDate != nil && len(categories) > 0 {
		categoryIDs := make([]uuid.UUID, len(categories))
		for i, cat := range categories {
			categoryIDs[i] = cat.ID
		}
		stats, err = uc.categoryRepo.GetTransactionStats(ctx, categoryIDs, *input.StartDate, *input.EndDate)
		if err != nil {
			// Log error but continue without stats
			stats = nil
		}
	}

	// Build output
	output := &ListCategoriesOutput{
		Categories: make([]*CategoryOutput, len(categories)),
	}

	for i, cat := range categories {
		categoryOutput := &CategoryOutput{
			ID:        cat.ID,
			Name:      cat.Name,
			Color:     cat.Color,
			Icon:      cat.Icon,
			OwnerType: cat.OwnerType,
			OwnerID:   cat.OwnerID,
			Type:      cat.Type,
			CreatedAt: cat.CreatedAt,
			UpdatedAt: cat.UpdatedAt,
		}

		// Add statistics if available
		if stats != nil {
			if catStats, ok := stats[cat.ID]; ok {
				categoryOutput.TransactionCount = catStats.TransactionCount
				categoryOutput.PeriodTotal = catStats.PeriodTotal
			}
		}

		output.Categories[i] = categoryOutput
	}

	return output, nil
}

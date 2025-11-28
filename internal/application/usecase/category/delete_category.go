// Package category contains category-related use cases.
package category

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// DeleteCategoryInput represents the input for category deletion.
type DeleteCategoryInput struct {
	CategoryID uuid.UUID
	OwnerType  entity.OwnerType
	OwnerID    uuid.UUID
}

// DeleteCategoryOutput represents the output of category deletion.
type DeleteCategoryOutput struct {
	Success bool
}

// DeleteCategoryUseCase handles category deletion logic.
type DeleteCategoryUseCase struct {
	categoryRepo adapter.CategoryRepository
}

// NewDeleteCategoryUseCase creates a new DeleteCategoryUseCase instance.
func NewDeleteCategoryUseCase(categoryRepo adapter.CategoryRepository) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{
		categoryRepo: categoryRepo,
	}
}

// Execute performs the category deletion.
func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, input DeleteCategoryInput) (*DeleteCategoryOutput, error) {
	// Find the existing category
	category, err := uc.categoryRepo.FindByID(ctx, input.CategoryID)
	if err != nil {
		if errors.Is(err, domainerror.ErrCategoryNotFound) {
			return nil, domainerror.NewCategoryError(
				domainerror.ErrCodeCategoryNotFound,
				"category not found",
				domainerror.ErrCategoryNotFound,
			)
		}
		return nil, fmt.Errorf("failed to find category: %w", err)
	}

	// Check if user is authorized to delete this category
	if category.OwnerType != input.OwnerType || category.OwnerID != input.OwnerID {
		return nil, domainerror.NewCategoryError(
			domainerror.ErrCodeNotAuthorizedCategory,
			"not authorized to delete this category",
			domainerror.ErrNotAuthorizedToModifyCategory,
		)
	}

	// Delete the category
	if err := uc.categoryRepo.Delete(ctx, input.CategoryID); err != nil {
		return nil, fmt.Errorf("failed to delete category: %w", err)
	}

	return &DeleteCategoryOutput{
		Success: true,
	}, nil
}

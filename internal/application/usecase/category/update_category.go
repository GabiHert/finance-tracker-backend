// Package category contains category-related use cases.
package category

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// UpdateCategoryInput represents the input for category update.
type UpdateCategoryInput struct {
	CategoryID uuid.UUID
	Name       *string // Optional
	Color      *string // Optional
	Icon       *string // Optional
	OwnerType  entity.OwnerType
	OwnerID    uuid.UUID
}

// UpdateCategoryOutput represents the output of category update.
type UpdateCategoryOutput struct {
	Category *entity.Category
}

// UpdateCategoryUseCase handles category update logic.
type UpdateCategoryUseCase struct {
	categoryRepo adapter.CategoryRepository
}

// NewUpdateCategoryUseCase creates a new UpdateCategoryUseCase instance.
func NewUpdateCategoryUseCase(categoryRepo adapter.CategoryRepository) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{
		categoryRepo: categoryRepo,
	}
}

// Execute performs the category update.
func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, input UpdateCategoryInput) (*UpdateCategoryOutput, error) {
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

	// Check if user is authorized to modify this category
	if category.OwnerType != input.OwnerType || category.OwnerID != input.OwnerID {
		return nil, domainerror.NewCategoryError(
			domainerror.ErrCodeNotAuthorizedCategory,
			"not authorized to modify this category",
			domainerror.ErrNotAuthorizedToModifyCategory,
		)
	}

	// Update name if provided
	if input.Name != nil {
		// Validate name length
		if len(*input.Name) > MaxCategoryNameLength {
			return nil, domainerror.NewCategoryError(
				domainerror.ErrCodeCategoryNameTooLong,
				fmt.Sprintf("category name must not exceed %d characters", MaxCategoryNameLength),
				domainerror.ErrCategoryNameTooLong,
			)
		}

		// Check if new name already exists for this owner (excluding current category)
		if *input.Name != category.Name {
			exists, err := uc.categoryRepo.ExistsByNameAndOwner(ctx, *input.Name, input.OwnerType, input.OwnerID)
			if err != nil {
				return nil, fmt.Errorf("failed to check category name existence: %w", err)
			}
			if exists {
				return nil, domainerror.NewCategoryError(
					domainerror.ErrCodeCategoryNameExists,
					"a category with this name already exists",
					domainerror.ErrCategoryNameExists,
				)
			}
		}

		category.Name = *input.Name
	}

	// Update color if provided
	if input.Color != nil {
		// Validate color format
		if *input.Color != "" && !isValidHexColor(*input.Color) {
			return nil, domainerror.NewCategoryError(
				domainerror.ErrCodeInvalidColorFormat,
				"color must be a valid hex format (#XXXXXX)",
				domainerror.ErrInvalidColorFormat,
			)
		}
		category.Color = *input.Color
	}

	// Update icon if provided
	if input.Icon != nil {
		category.Icon = *input.Icon
	}

	// Update timestamp
	category.UpdatedAt = time.Now().UTC()

	// Save updated category
	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &UpdateCategoryOutput{
		Category: category,
	}, nil
}

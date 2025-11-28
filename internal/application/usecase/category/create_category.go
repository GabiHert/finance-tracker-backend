// Package category contains category-related use cases.
package category

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
	// MaxCategoryNameLength is the maximum allowed length for category names.
	MaxCategoryNameLength = 50
	// MaxIconLength is the maximum allowed length for icon names.
	MaxIconLength = 50
)

// hexColorRegex is compiled once at package level for performance.
var hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

// CreateCategoryInput represents the input for category creation.
type CreateCategoryInput struct {
	Name      string
	Color     string // Optional, defaults to DefaultCategoryColor
	Icon      string // Optional, defaults to DefaultCategoryIcon
	OwnerType entity.OwnerType
	OwnerID   uuid.UUID
	Type      entity.CategoryType
}

// CreateCategoryOutput represents the output of category creation.
type CreateCategoryOutput struct {
	Category *entity.Category
}

// CreateCategoryUseCase handles category creation logic.
type CreateCategoryUseCase struct {
	categoryRepo adapter.CategoryRepository
}

// NewCreateCategoryUseCase creates a new CreateCategoryUseCase instance.
func NewCreateCategoryUseCase(categoryRepo adapter.CategoryRepository) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{
		categoryRepo: categoryRepo,
	}
}

// Execute performs the category creation.
func (uc *CreateCategoryUseCase) Execute(ctx context.Context, input CreateCategoryInput) (*CreateCategoryOutput, error) {
	// Validate name length
	if len(input.Name) > MaxCategoryNameLength {
		return nil, domainerror.NewCategoryError(
			domainerror.ErrCodeCategoryNameTooLong,
			fmt.Sprintf("category name must not exceed %d characters", MaxCategoryNameLength),
			domainerror.ErrCategoryNameTooLong,
		)
	}

	// Validate color format if provided
	if input.Color != "" && !isValidHexColor(input.Color) {
		return nil, domainerror.NewCategoryError(
			domainerror.ErrCodeInvalidColorFormat,
			"color must be a valid hex format (#XXXXXX)",
			domainerror.ErrInvalidColorFormat,
		)
	}

	// Apply default values for optional fields (Application layer responsibility)
	color := input.Color
	if color == "" {
		color = entity.DefaultCategoryColor
	}
	icon := input.Icon
	if icon == "" {
		icon = entity.DefaultCategoryIcon
	}

	// Validate category type
	if !isValidCategoryType(input.Type) {
		return nil, domainerror.NewCategoryError(
			domainerror.ErrCodeInvalidCategoryType,
			"category type must be 'expense' or 'income'",
			domainerror.ErrInvalidCategoryType,
		)
	}

	// Validate owner type
	if !isValidOwnerType(input.OwnerType) {
		return nil, domainerror.NewCategoryError(
			domainerror.ErrCodeInvalidOwnerType,
			"owner type must be 'user' or 'group'",
			domainerror.ErrInvalidOwnerType,
		)
	}

	// Check if category name already exists for this owner
	exists, err := uc.categoryRepo.ExistsByNameAndOwner(ctx, input.Name, input.OwnerType, input.OwnerID)
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

	// Create category entity with defaulted values
	category := entity.NewCategory(
		input.Name,
		color,
		icon,
		input.OwnerType,
		input.OwnerID,
		input.Type,
	)

	// Save category to database
	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &CreateCategoryOutput{
		Category: category,
	}, nil
}

// isValidHexColor validates hex color format (#XXXXXX or #XXX).
func isValidHexColor(color string) bool {
	return hexColorRegex.MatchString(color)
}

// isValidCategoryType validates the category type.
func isValidCategoryType(categoryType entity.CategoryType) bool {
	return categoryType == entity.CategoryTypeExpense || categoryType == entity.CategoryTypeIncome
}

// isValidOwnerType validates the owner type.
func isValidOwnerType(ownerType entity.OwnerType) bool {
	return ownerType == entity.OwnerTypeUser || ownerType == entity.OwnerTypeGroup
}

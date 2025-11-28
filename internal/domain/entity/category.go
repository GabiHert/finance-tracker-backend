// Package entity defines the core business entities for the domain layer.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// OwnerType represents the type of owner for a category.
type OwnerType string

const (
	OwnerTypeUser  OwnerType = "user"
	OwnerTypeGroup OwnerType = "group"
)

// CategoryType represents the type of category (expense or income).
type CategoryType string

const (
	CategoryTypeExpense CategoryType = "expense"
	CategoryTypeIncome  CategoryType = "income"
)

// DefaultCategoryColor is the default color for categories.
const DefaultCategoryColor = "#6366F1"

// DefaultCategoryIcon is the default icon for categories.
const DefaultCategoryIcon = "tag"

// Category represents a transaction category in the Finance Tracker system.
type Category struct {
	ID        uuid.UUID
	Name      string
	Color     string
	Icon      string
	OwnerType OwnerType
	OwnerID   uuid.UUID
	Type      CategoryType
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // Soft-delete support
}

// NewCategory creates a new Category entity.
// Note: Defaulting logic for color and icon should be applied in the Application layer (UseCase)
// before calling this constructor.
func NewCategory(name, color, icon string, ownerType OwnerType, ownerID uuid.UUID, categoryType CategoryType) *Category {
	now := time.Now().UTC()

	return &Category{
		ID:        uuid.New(),
		Name:      name,
		Color:     color,
		Icon:      icon,
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Type:      categoryType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// CategoryWithStats represents a category with transaction statistics.
type CategoryWithStats struct {
	Category         *Category
	TransactionCount int
	PeriodTotal      float64
}

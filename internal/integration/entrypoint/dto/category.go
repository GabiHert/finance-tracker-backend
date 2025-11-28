// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"time"

	"github.com/finance-tracker/backend/internal/application/usecase/category"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// CreateCategoryRequest represents the request body for category creation.
type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=50"`
	Color string `json:"color,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Type  string `json:"type" binding:"required,oneof=expense income"`
}

// UpdateCategoryRequest represents the request body for category update.
type UpdateCategoryRequest struct {
	Name  *string `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Color *string `json:"color,omitempty"`
	Icon  *string `json:"icon,omitempty"`
}

// CategoryResponse represents a single category in API responses.
type CategoryResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Color            string    `json:"color"`
	Icon             string    `json:"icon"`
	OwnerType        string    `json:"owner_type"`
	OwnerID          string    `json:"owner_id"`
	Type             string    `json:"type"`
	TransactionCount int       `json:"transaction_count"`
	PeriodTotal      float64   `json:"period_total"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CategoryListResponse represents the response for listing categories.
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

// ToCategoryResponse converts a domain Category entity to a CategoryResponse DTO.
func ToCategoryResponse(cat *entity.Category) CategoryResponse {
	return CategoryResponse{
		ID:               cat.ID.String(),
		Name:             cat.Name,
		Color:            cat.Color,
		Icon:             cat.Icon,
		OwnerType:        string(cat.OwnerType),
		OwnerID:          cat.OwnerID.String(),
		Type:             string(cat.Type),
		TransactionCount: 0,
		PeriodTotal:      0,
		CreatedAt:        cat.CreatedAt,
		UpdatedAt:        cat.UpdatedAt,
	}
}

// ToCategoryResponseWithStats converts a CategoryOutput to a CategoryResponse DTO.
func ToCategoryResponseWithStats(output *category.CategoryOutput) CategoryResponse {
	return CategoryResponse{
		ID:               output.ID.String(),
		Name:             output.Name,
		Color:            output.Color,
		Icon:             output.Icon,
		OwnerType:        string(output.OwnerType),
		OwnerID:          output.OwnerID.String(),
		Type:             string(output.Type),
		TransactionCount: output.TransactionCount,
		PeriodTotal:      output.PeriodTotal,
		CreatedAt:        output.CreatedAt,
		UpdatedAt:        output.UpdatedAt,
	}
}

// ToCategoryListResponse converts a list of CategoryOutput to CategoryListResponse.
func ToCategoryListResponse(outputs []*category.CategoryOutput) CategoryListResponse {
	categories := make([]CategoryResponse, len(outputs))
	for i, output := range outputs {
		categories[i] = ToCategoryResponseWithStats(output)
	}
	return CategoryListResponse{
		Categories: categories,
	}
}

// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/usecase/category"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// CategoryController handles category endpoints.
type CategoryController struct {
	listUseCase   *category.ListCategoriesUseCase
	createUseCase *category.CreateCategoryUseCase
	updateUseCase *category.UpdateCategoryUseCase
	deleteUseCase *category.DeleteCategoryUseCase
}

// NewCategoryController creates a new category controller instance.
func NewCategoryController(
	listUseCase *category.ListCategoriesUseCase,
	createUseCase *category.CreateCategoryUseCase,
	updateUseCase *category.UpdateCategoryUseCase,
	deleteUseCase *category.DeleteCategoryUseCase,
) *CategoryController {
	return &CategoryController{
		listUseCase:   listUseCase,
		createUseCase: createUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
	}
}

// List handles GET /categories requests.
func (c *CategoryController) List(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse query parameters
	input := category.ListCategoriesInput{
		OwnerType: entity.OwnerTypeUser,
		OwnerID:   userID,
	}

	// Filter by category type if provided
	if categoryType := ctx.Query("type"); categoryType != "" {
		catType := entity.CategoryType(categoryType)
		input.CategoryType = &catType
	}

	// Parse date range for statistics
	if startDateStr := ctx.Query("startDate"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			input.StartDate = &startDate
		}
	}
	if endDateStr := ctx.Query("endDate"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			input.EndDate = &endDate
		}
	}

	// Execute use case
	output, err := c.listUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve categories",
		})
		return
	}

	// Build response
	response := dto.ToCategoryListResponse(output.Categories)
	ctx.JSON(http.StatusOK, response)
}

// Create handles POST /categories requests.
func (c *CategoryController) Create(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse request body
	var req dto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingCategoryFields),
		})
		return
	}

	// Build input
	input := category.CreateCategoryInput{
		Name:      req.Name,
		Color:     req.Color,
		Icon:      req.Icon,
		OwnerType: entity.OwnerTypeUser,
		OwnerID:   userID,
		Type:      entity.CategoryType(req.Type),
	}

	// Execute use case
	output, err := c.createUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryError(ctx, err)
		return
	}

	// Build response
	response := dto.ToCategoryResponse(output.Category)
	ctx.JSON(http.StatusCreated, response)
}

// Update handles PATCH /categories/:id requests.
func (c *CategoryController) Update(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse category ID from URL
	categoryIDStr := ctx.Param("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid category ID format",
		})
		return
	}

	// Parse request body
	var req dto.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Build input
	input := category.UpdateCategoryInput{
		CategoryID: categoryID,
		Name:       req.Name,
		Color:      req.Color,
		Icon:       req.Icon,
		OwnerType:  entity.OwnerTypeUser,
		OwnerID:    userID,
	}

	// Execute use case
	output, err := c.updateUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryError(ctx, err)
		return
	}

	// Build response
	response := dto.ToCategoryResponse(output.Category)
	ctx.JSON(http.StatusOK, response)
}

// Delete handles DELETE /categories/:id requests.
func (c *CategoryController) Delete(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse category ID from URL
	categoryIDStr := ctx.Param("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid category ID format",
		})
		return
	}

	// Build input
	input := category.DeleteCategoryInput{
		CategoryID: categoryID,
		OwnerType:  entity.OwnerTypeUser,
		OwnerID:    userID,
	}

	// Execute use case
	_, err = c.deleteUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryError(ctx, err)
		return
	}

	// Return no content on success
	ctx.Status(http.StatusNoContent)
}

// handleCategoryError handles category errors and returns appropriate HTTP responses.
func (c *CategoryController) handleCategoryError(ctx *gin.Context, err error) {
	var catErr *domainerror.CategoryError
	if errors.As(err, &catErr) {
		statusCode := c.getStatusCodeForCategoryError(catErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: catErr.Message,
			Code:  string(catErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForCategoryError maps category error codes to HTTP status codes.
func (c *CategoryController) getStatusCodeForCategoryError(code domainerror.CategoryErrorCode) int {
	switch code {
	case domainerror.ErrCodeCategoryNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeCategoryNameExists:
		return http.StatusConflict
	case domainerror.ErrCodeNotAuthorizedCategory:
		return http.StatusForbidden
	case domainerror.ErrCodeCategoryNameTooLong,
		domainerror.ErrCodeInvalidColorFormat,
		domainerror.ErrCodeInvalidOwnerType,
		domainerror.ErrCodeInvalidCategoryType,
		domainerror.ErrCodeMissingCategoryFields:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

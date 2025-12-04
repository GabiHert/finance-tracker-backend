// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	categoryrule "github.com/finance-tracker/backend/internal/application/usecase/category_rule"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// CategoryRuleController handles category rule endpoints.
type CategoryRuleController struct {
	listUseCase    *categoryrule.ListCategoryRulesUseCase
	createUseCase  *categoryrule.CreateCategoryRuleUseCase
	updateUseCase  *categoryrule.UpdateCategoryRuleUseCase
	deleteUseCase  *categoryrule.DeleteCategoryRuleUseCase
	reorderUseCase *categoryrule.ReorderCategoryRulesUseCase
	testUseCase    *categoryrule.TestPatternUseCase
}

// NewCategoryRuleController creates a new category rule controller instance.
func NewCategoryRuleController(
	listUseCase *categoryrule.ListCategoryRulesUseCase,
	createUseCase *categoryrule.CreateCategoryRuleUseCase,
	updateUseCase *categoryrule.UpdateCategoryRuleUseCase,
	deleteUseCase *categoryrule.DeleteCategoryRuleUseCase,
	reorderUseCase *categoryrule.ReorderCategoryRulesUseCase,
	testUseCase *categoryrule.TestPatternUseCase,
) *CategoryRuleController {
	return &CategoryRuleController{
		listUseCase:    listUseCase,
		createUseCase:  createUseCase,
		updateUseCase:  updateUseCase,
		deleteUseCase:  deleteUseCase,
		reorderUseCase: reorderUseCase,
		testUseCase:    testUseCase,
	}
}

// List handles GET /category-rules requests.
func (c *CategoryRuleController) List(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Build input
	input := categoryrule.ListCategoryRulesInput{
		OwnerType:  entity.OwnerTypeUser,
		OwnerID:    userID,
		ActiveOnly: false,
	}

	// Check for activeOnly query param
	if ctx.Query("active_only") == "true" {
		input.ActiveOnly = true
	}

	// Execute use case
	output, err := c.listUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve category rules",
		})
		return
	}

	// Build response
	response := dto.ToCategoryRuleListResponse(output.Rules)
	ctx.JSON(http.StatusOK, response)
}

// Create handles POST /category-rules requests.
func (c *CategoryRuleController) Create(ctx *gin.Context) {
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
	var req dto.CreateCategoryRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingRuleFields),
		})
		return
	}

	// Parse category ID
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid category ID format",
		})
		return
	}

	// Build input
	input := categoryrule.CreateCategoryRuleInput{
		Pattern:    req.Pattern,
		CategoryID: categoryID,
		Priority:   req.Priority,
		OwnerType:  entity.OwnerTypeUser,
		OwnerID:    userID,
	}

	// Execute use case
	output, err := c.createUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryRuleError(ctx, err)
		return
	}

	// Build response
	response := dto.ToCategoryRuleResponse(output.Rule)
	response.TransactionsUpdated = output.TransactionsUpdated
	ctx.JSON(http.StatusCreated, response)
}

// Update handles PATCH /category-rules/:id requests.
func (c *CategoryRuleController) Update(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse rule ID from URL
	ruleIDStr := ctx.Param("id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid rule ID format",
		})
		return
	}

	// Parse request body
	var req dto.UpdateCategoryRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Build input
	input := categoryrule.UpdateCategoryRuleInput{
		RuleID:    ruleID,
		Pattern:   req.Pattern,
		Priority:  req.Priority,
		IsActive:  req.IsActive,
		OwnerType: entity.OwnerTypeUser,
		OwnerID:   userID,
	}

	// Parse category ID if provided
	if req.CategoryID != nil {
		categoryID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid category ID format",
			})
			return
		}
		input.CategoryID = &categoryID
	}

	// Execute use case
	output, err := c.updateUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryRuleError(ctx, err)
		return
	}

	// Build response
	response := dto.ToCategoryRuleResponse(output.Rule)
	ctx.JSON(http.StatusOK, response)
}

// Delete handles DELETE /category-rules/:id requests.
func (c *CategoryRuleController) Delete(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse rule ID from URL
	ruleIDStr := ctx.Param("id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid rule ID format",
		})
		return
	}

	// Build input
	input := categoryrule.DeleteCategoryRuleInput{
		RuleID:    ruleID,
		OwnerType: entity.OwnerTypeUser,
		OwnerID:   userID,
	}

	// Execute use case
	_, err = c.deleteUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryRuleError(ctx, err)
		return
	}

	// Return no content on success
	ctx.Status(http.StatusNoContent)
}

// Reorder handles PATCH /category-rules/reorder requests.
func (c *CategoryRuleController) Reorder(ctx *gin.Context) {
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
	var req dto.ReorderCategoryRulesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingRuleFields),
		})
		return
	}

	// Build input
	order := make([]categoryrule.RulePriorityInput, len(req.Order))
	for i, item := range req.Order {
		ruleID, err := uuid.Parse(item.ID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid rule ID format",
			})
			return
		}
		order[i] = categoryrule.RulePriorityInput{
			ID:       ruleID,
			Priority: item.Priority,
		}
	}

	input := categoryrule.ReorderCategoryRulesInput{
		Order:     order,
		OwnerType: entity.OwnerTypeUser,
		OwnerID:   userID,
	}

	// Execute use case
	output, err := c.reorderUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryRuleError(ctx, err)
		return
	}

	// Build response
	response := dto.ToCategoryRuleListResponse(output.Rules)
	ctx.JSON(http.StatusOK, response)
}

// TestPattern handles POST /category-rules/test requests.
func (c *CategoryRuleController) TestPattern(ctx *gin.Context) {
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
	var req dto.TestPatternRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingRuleFields),
		})
		return
	}

	// Build input
	input := categoryrule.TestPatternInput{
		Pattern:   req.Pattern,
		OwnerType: entity.OwnerTypeUser,
		OwnerID:   userID,
	}

	// Execute use case
	output, err := c.testUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCategoryRuleError(ctx, err)
		return
	}

	// Build response
	response := dto.ToTestPatternResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// handleCategoryRuleError handles category rule errors and returns appropriate HTTP responses.
func (c *CategoryRuleController) handleCategoryRuleError(ctx *gin.Context, err error) {
	var ruleErr *domainerror.CategoryRuleError
	if errors.As(err, &ruleErr) {
		statusCode := c.getStatusCodeForCategoryRuleError(ruleErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: ruleErr.Message,
			Code:  string(ruleErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForCategoryRuleError maps category rule error codes to HTTP status codes.
func (c *CategoryRuleController) getStatusCodeForCategoryRuleError(code domainerror.CategoryRuleErrorCode) int {
	switch code {
	case domainerror.ErrCodeCategoryRuleNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeCategoryRulePatternExists:
		return http.StatusConflict
	case domainerror.ErrCodeNotAuthorizedRule:
		return http.StatusForbidden
	case domainerror.ErrCodeCategoryNotFoundForRule:
		return http.StatusNotFound
	case domainerror.ErrCodeInvalidPattern,
		domainerror.ErrCodePatternTooLong,
		domainerror.ErrCodeMissingRuleFields,
		domainerror.ErrCodeInvalidPriority,
		domainerror.ErrCodeRuleOwnerTypeMismatch:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

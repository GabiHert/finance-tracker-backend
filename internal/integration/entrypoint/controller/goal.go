// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/usecase/goal"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// GoalController handles goal endpoints.
type GoalController struct {
	listUseCase   *goal.ListGoalsUseCase
	createUseCase *goal.CreateGoalUseCase
	getUseCase    *goal.GetGoalUseCase
	updateUseCase *goal.UpdateGoalUseCase
	deleteUseCase *goal.DeleteGoalUseCase
}

// NewGoalController creates a new goal controller instance.
func NewGoalController(
	listUseCase *goal.ListGoalsUseCase,
	createUseCase *goal.CreateGoalUseCase,
	getUseCase *goal.GetGoalUseCase,
	updateUseCase *goal.UpdateGoalUseCase,
	deleteUseCase *goal.DeleteGoalUseCase,
) *GoalController {
	return &GoalController{
		listUseCase:   listUseCase,
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
	}
}

// List handles GET /goals requests.
func (c *GoalController) List(ctx *gin.Context) {
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
	input := goal.ListGoalsInput{
		UserID: userID,
	}

	// Execute use case
	output, err := c.listUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve goals",
		})
		return
	}

	// Build response
	response := dto.ToGoalListResponse(output.Goals)
	ctx.JSON(http.StatusOK, response)
}

// Create handles POST /goals requests.
func (c *GoalController) Create(ctx *gin.Context) {
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
	var req dto.CreateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
			Code:  string(domainerror.ErrCodeMissingGoalFields),
		})
		return
	}

	// Parse category ID
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid category ID format",
			Code:  string(domainerror.ErrCodeMissingGoalFields),
		})
		return
	}

	// Build input
	input := goal.CreateGoalInput{
		UserID:        userID,
		CategoryID:    categoryID,
		LimitAmount:   req.LimitAmount,
		AlertOnExceed: req.AlertOnExceed,
	}

	// Convert period if provided
	if req.Period != nil {
		period := entity.GoalPeriod(*req.Period)
		input.Period = &period
	}

	// Execute use case
	output, err := c.createUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGoalError(ctx, err)
		return
	}

	// Build response
	response := dto.ToGoalResponse(output.Goal)
	ctx.JSON(http.StatusCreated, response)
}

// Get handles GET /goals/:id requests.
func (c *GoalController) Get(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse goal ID from URL
	goalIDStr := ctx.Param("id")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid goal ID format",
		})
		return
	}

	// Build input
	input := goal.GetGoalInput{
		GoalID: goalID,
		UserID: userID,
	}

	// Execute use case
	output, err := c.getUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGoalError(ctx, err)
		return
	}

	// Build response
	response := dto.ToGoalResponse(output.Goal)
	response.CurrentAmount = output.CurrentAmount
	if output.Category != nil {
		catResponse := dto.ToCategoryResponse(output.Category)
		response.Category = &catResponse
	}
	ctx.JSON(http.StatusOK, response)
}

// Update handles PATCH /goals/:id requests.
func (c *GoalController) Update(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse goal ID from URL
	goalIDStr := ctx.Param("id")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid goal ID format",
		})
		return
	}

	// Parse request body
	var req dto.UpdateGoalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Build input
	input := goal.UpdateGoalInput{
		GoalID:        goalID,
		UserID:        userID,
		LimitAmount:   req.LimitAmount,
		AlertOnExceed: req.AlertOnExceed,
	}

	// Convert period if provided
	if req.Period != nil {
		period := entity.GoalPeriod(*req.Period)
		input.Period = &period
	}

	// Execute use case
	output, err := c.updateUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGoalError(ctx, err)
		return
	}

	// Build response
	response := dto.ToGoalResponse(output.Goal)
	ctx.JSON(http.StatusOK, response)
}

// Delete handles DELETE /goals/:id requests.
func (c *GoalController) Delete(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse goal ID from URL
	goalIDStr := ctx.Param("id")
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid goal ID format",
		})
		return
	}

	// Build input
	input := goal.DeleteGoalInput{
		GoalID: goalID,
		UserID: userID,
	}

	// Execute use case
	_, err = c.deleteUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleGoalError(ctx, err)
		return
	}

	// Return no content on success
	ctx.Status(http.StatusNoContent)
}

// handleGoalError handles goal errors and returns appropriate HTTP responses.
func (c *GoalController) handleGoalError(ctx *gin.Context, err error) {
	var goalErr *domainerror.GoalError
	if errors.As(err, &goalErr) {
		statusCode := c.getStatusCodeForGoalError(goalErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: goalErr.Message,
			Code:  string(goalErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForGoalError maps goal error codes to HTTP status codes.
func (c *GoalController) getStatusCodeForGoalError(code domainerror.GoalErrorCode) int {
	switch code {
	case domainerror.ErrCodeGoalNotFound, domainerror.ErrCodeGoalCategoryNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeGoalAlreadyExists:
		return http.StatusConflict
	case domainerror.ErrCodeUnauthorizedGoalAccess, domainerror.ErrCodeCategoryDoesNotBelongUser:
		return http.StatusForbidden
	case domainerror.ErrCodeInvalidLimitAmount,
		domainerror.ErrCodeInvalidGoalPeriod,
		domainerror.ErrCodeMissingGoalFields:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

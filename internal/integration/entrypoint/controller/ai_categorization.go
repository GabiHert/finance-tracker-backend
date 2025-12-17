// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	aicategorization "github.com/finance-tracker/backend/internal/application/usecase/ai_categorization"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// AiCategorizationController handles AI categorization endpoints.
type AiCategorizationController struct {
	getStatusUseCase      *aicategorization.GetStatusUseCase
	startUseCase          *aicategorization.StartCategorizationUseCase
	getSuggestionsUseCase *aicategorization.GetSuggestionsUseCase
	approveUseCase        *aicategorization.ApproveSuggestionUseCase
	rejectUseCase         *aicategorization.RejectSuggestionUseCase
	clearUseCase          *aicategorization.ClearSuggestionsUseCase
}

// NewAiCategorizationController creates a new AI categorization controller instance.
func NewAiCategorizationController(
	getStatusUseCase *aicategorization.GetStatusUseCase,
	startUseCase *aicategorization.StartCategorizationUseCase,
	getSuggestionsUseCase *aicategorization.GetSuggestionsUseCase,
	approveUseCase *aicategorization.ApproveSuggestionUseCase,
	rejectUseCase *aicategorization.RejectSuggestionUseCase,
	clearUseCase *aicategorization.ClearSuggestionsUseCase,
) *AiCategorizationController {
	return &AiCategorizationController{
		getStatusUseCase:      getStatusUseCase,
		startUseCase:          startUseCase,
		getSuggestionsUseCase: getSuggestionsUseCase,
		approveUseCase:        approveUseCase,
		rejectUseCase:         rejectUseCase,
		clearUseCase:          clearUseCase,
	}
}

// GetStatus handles GET /ai/categorization/status requests.
func (c *AiCategorizationController) GetStatus(ctx *gin.Context) {
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
	input := aicategorization.GetStatusInput{
		UserID: userID,
	}

	// Execute use case
	output, err := c.getStatusUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAICategorizationError(ctx, err)
		return
	}

	// Build response
	response := dto.ToCategorizationStatusResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// Start handles POST /ai/categorization/start requests.
func (c *AiCategorizationController) Start(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse request body (optional, currently empty)
	var req dto.StartCategorizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Ignore binding errors for empty request body
	}

	// Build input
	input := aicategorization.StartCategorizationInput{
		UserID: userID,
	}

	// Execute use case
	output, err := c.startUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAICategorizationError(ctx, err)
		return
	}

	// Build response - 202 Accepted for async operation
	response := dto.ToStartCategorizationResponse(output)
	ctx.JSON(http.StatusAccepted, response)
}

// GetSuggestions handles GET /ai/categorization/suggestions requests.
func (c *AiCategorizationController) GetSuggestions(ctx *gin.Context) {
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
	input := aicategorization.GetSuggestionsInput{
		UserID: userID,
	}

	// Execute use case
	output, err := c.getSuggestionsUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAICategorizationError(ctx, err)
		return
	}

	// Build response
	response := dto.ToSuggestionsListResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// ApproveSuggestion handles POST /ai/categorization/suggestions/:id/approve requests.
func (c *AiCategorizationController) ApproveSuggestion(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse suggestion ID from URL
	suggestionIDStr := ctx.Param("id")
	suggestionID, err := uuid.Parse(suggestionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid suggestion ID format",
		})
		return
	}

	// Parse request body (optional, currently empty)
	var req dto.ApproveSuggestionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Ignore binding errors for empty request body
	}

	// Build input
	input := aicategorization.ApproveSuggestionInput{
		SuggestionID: suggestionID,
		UserID:       userID,
	}

	// Execute use case
	output, err := c.approveUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAICategorizationError(ctx, err)
		return
	}

	// Build response
	response := dto.ToApproveSuggestionResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// RejectSuggestion handles POST /ai/categorization/suggestions/:id/reject requests.
func (c *AiCategorizationController) RejectSuggestion(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse suggestion ID from URL
	suggestionIDStr := ctx.Param("id")
	suggestionID, err := uuid.Parse(suggestionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid suggestion ID format",
		})
		return
	}

	// Parse request body
	var req dto.RejectSuggestionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: action must be 'skip' or 'retry'",
		})
		return
	}

	// Build input
	input := aicategorization.RejectSuggestionInput{
		SuggestionID: suggestionID,
		UserID:       userID,
		Action:       aicategorization.RejectAction(req.Action),
		RetryReason:  req.RetryReason,
	}

	// Execute use case
	output, err := c.rejectUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAICategorizationError(ctx, err)
		return
	}

	// Build response
	response := dto.ToRejectSuggestionResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// ClearSuggestions handles DELETE /ai/categorization/suggestions requests.
func (c *AiCategorizationController) ClearSuggestions(ctx *gin.Context) {
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
	input := aicategorization.ClearSuggestionsInput{
		UserID: userID,
	}

	// Execute use case
	output, err := c.clearUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAICategorizationError(ctx, err)
		return
	}

	// Build response
	response := dto.ToClearSuggestionsResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// handleAICategorizationError handles AI categorization errors and returns appropriate HTTP responses.
func (c *AiCategorizationController) handleAICategorizationError(ctx *gin.Context, err error) {
	var aiErr *domainerror.AISuggestionError
	if errors.As(err, &aiErr) {
		statusCode := c.getStatusCodeForAIError(aiErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: aiErr.Message,
			Code:  string(aiErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForAIError maps AI categorization error codes to HTTP status codes.
func (c *AiCategorizationController) getStatusCodeForAIError(code domainerror.AISuggestionErrorCode) int {
	switch code {
	case domainerror.ErrCodeAISuggestionNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeAIAlreadyProcessing:
		return http.StatusConflict
	case domainerror.ErrCodeAINoUncategorized:
		return http.StatusBadRequest
	case domainerror.ErrCodeAIPatternConflict:
		return http.StatusConflict
	case domainerror.ErrCodeAIInvalidMatchType,
		domainerror.ErrCodeAIEmptyKeyword,
		domainerror.ErrCodeAIInvalidAction:
		return http.StatusBadRequest
	case domainerror.ErrCodeAISuggestionAlreadyProcessed:
		return http.StatusConflict
	case domainerror.ErrCodeAIServiceError,
		domainerror.ErrCodeAIRetryFailed,
		domainerror.ErrCodeAIInvalidConfig:
		return http.StatusInternalServerError
	case domainerror.ErrCodeAIRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/usecase/dashboard"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// DashboardController handles dashboard endpoints.
type DashboardController struct {
	getCategoryTrendsUseCase       *dashboard.GetCategoryTrendsUseCase
	getDataRangeUseCase            *dashboard.GetDataRangeUseCase
	getTrendsUseCase               *dashboard.GetTrendsUseCase
	getCategoryBreakdownUseCase    *dashboard.GetCategoryBreakdownUseCase
	getPeriodTransactionsUseCase   *dashboard.GetPeriodTransactionsUseCase
}

// NewDashboardController creates a new dashboard controller instance.
func NewDashboardController(
	getCategoryTrendsUseCase *dashboard.GetCategoryTrendsUseCase,
	getDataRangeUseCase *dashboard.GetDataRangeUseCase,
	getTrendsUseCase *dashboard.GetTrendsUseCase,
	getCategoryBreakdownUseCase *dashboard.GetCategoryBreakdownUseCase,
	getPeriodTransactionsUseCase *dashboard.GetPeriodTransactionsUseCase,
) *DashboardController {
	return &DashboardController{
		getCategoryTrendsUseCase:       getCategoryTrendsUseCase,
		getDataRangeUseCase:            getDataRangeUseCase,
		getTrendsUseCase:               getTrendsUseCase,
		getCategoryBreakdownUseCase:    getCategoryBreakdownUseCase,
		getPeriodTransactionsUseCase:   getPeriodTransactionsUseCase,
	}
}

// GetCategoryTrends handles GET /dashboard/category-trends requests.
func (c *DashboardController) GetCategoryTrends(ctx *gin.Context) {
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
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	granularity := ctx.Query("granularity")
	topCategoriesStr := ctx.DefaultQuery("top_categories", "8")

	// Validate required parameters
	if startDateStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "start_date is required",
			Code:  string(domainerror.ErrCodeMissingStartDate),
		})
		return
	}

	if endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "end_date is required",
			Code:  string(domainerror.ErrCodeMissingEndDate),
		})
		return
	}

	if granularity == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "granularity is required",
			Code:  string(domainerror.ErrCodeMissingGranularity),
		})
		return
	}

	// Validate and parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid start_date format, expected YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidDateFormat),
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid end_date format, expected YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidDateFormat),
		})
		return
	}

	// Validate granularity
	gran := dashboard.Granularity(granularity)
	if gran != dashboard.GranularityDaily &&
		gran != dashboard.GranularityWeekly &&
		gran != dashboard.GranularityMonthly {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Granularity must be: daily, weekly, or monthly",
			Code:  string(domainerror.ErrCodeInvalidGranularity),
		})
		return
	}

	// Parse top categories
	topCategories, err := strconv.Atoi(topCategoriesStr)
	if err != nil || topCategories <= 0 {
		topCategories = 8
	}

	// Execute use case
	input := dashboard.GetCategoryTrendsInput{
		UserID:        userID,
		StartDate:     startDate,
		EndDate:       endDate,
		Granularity:   gran,
		TopCategories: topCategories,
	}

	output, err := c.getCategoryTrendsUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleDashboardError(ctx, err)
		return
	}

	// Transform to response DTO
	response := dto.ToCategoryTrendsResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// handleDashboardError handles dashboard errors and returns appropriate HTTP responses.
func (c *DashboardController) handleDashboardError(ctx *gin.Context, err error) {
	var dashErr *domainerror.DashboardError
	if errors.As(err, &dashErr) {
		statusCode := c.getStatusCodeForDashboardError(dashErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: dashErr.Message,
			Code:  string(dashErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForDashboardError maps dashboard error codes to HTTP status codes.
func (c *DashboardController) getStatusCodeForDashboardError(code domainerror.DashboardErrorCode) int {
	switch code {
	case domainerror.ErrCodeMissingStartDate,
		domainerror.ErrCodeMissingEndDate,
		domainerror.ErrCodeInvalidDateRange,
		domainerror.ErrCodeInvalidGranularity,
		domainerror.ErrCodeMissingGranularity,
		domainerror.ErrCodeInvalidDateFormat:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// GetDataRange handles GET /dashboard/data-range requests.
func (c *DashboardController) GetDataRange(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Execute use case
	input := dashboard.GetDataRangeInput{
		UserID: userID,
	}

	output, err := c.getDataRangeUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleDashboardError(ctx, err)
		return
	}

	// Transform to response DTO
	response := dto.ToDataRangeResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// GetTrends handles GET /dashboard/trends requests.
func (c *DashboardController) GetTrends(ctx *gin.Context) {
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
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	granularity := ctx.Query("granularity")

	// Validate required parameters
	if startDateStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "start_date is required",
			Code:  string(domainerror.ErrCodeMissingStartDate),
		})
		return
	}

	if endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "end_date is required",
			Code:  string(domainerror.ErrCodeMissingEndDate),
		})
		return
	}

	if granularity == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "granularity is required",
			Code:  string(domainerror.ErrCodeMissingGranularity),
		})
		return
	}

	// Validate and parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid start_date format, expected YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidDateFormat),
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid end_date format, expected YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidDateFormat),
		})
		return
	}

	// Validate granularity
	gran := dashboard.Granularity(granularity)
	if gran != dashboard.GranularityWeekly &&
		gran != dashboard.GranularityMonthly &&
		gran != dashboard.GranularityQuarterly {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Granularity must be: weekly, monthly, or quarterly",
			Code:  string(domainerror.ErrCodeInvalidGranularity),
		})
		return
	}

	// Execute use case
	input := dashboard.GetTrendsInput{
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: gran,
	}

	output, err := c.getTrendsUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleDashboardError(ctx, err)
		return
	}

	// Transform to response DTO
	response := dto.ToTrendsResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// GetCategoryBreakdown handles GET /dashboard/category-breakdown requests.
func (c *DashboardController) GetCategoryBreakdown(ctx *gin.Context) {
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
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	// Validate required parameters
	if startDateStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "start_date is required",
			Code:  string(domainerror.ErrCodeMissingStartDate),
		})
		return
	}

	if endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "end_date is required",
			Code:  string(domainerror.ErrCodeMissingEndDate),
		})
		return
	}

	// Validate and parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid start_date format, expected YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidDateFormat),
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid end_date format, expected YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidDateFormat),
		})
		return
	}

	// Execute use case
	input := dashboard.GetCategoryBreakdownInput{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	output, err := c.getCategoryBreakdownUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleDashboardError(ctx, err)
		return
	}

	// Transform to response DTO
	response := dto.ToCategoryBreakdownResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// GetPeriodTransactions handles GET /dashboard/period-transactions requests.
func (c *DashboardController) GetPeriodTransactions(ctx *gin.Context) {
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
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	categoryIDStr := ctx.Query("category_id")
	limitStr := ctx.DefaultQuery("limit", "50")
	offsetStr := ctx.DefaultQuery("offset", "0")

	// Validate required parameters
	if startDateStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "start_date is required",
			Code:  string(domainerror.ErrCodeMissingStartDate),
		})
		return
	}

	if endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "end_date is required",
			Code:  string(domainerror.ErrCodeMissingEndDate),
		})
		return
	}

	// Validate and parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid start_date format, expected YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidDateFormat),
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid end_date format, expected YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidDateFormat),
		})
		return
	}

	// Parse optional category ID
	var categoryID *uuid.UUID
	if categoryIDStr != "" {
		catID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid category_id format",
			})
			return
		}
		categoryID = &catID
	}

	// Parse pagination
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Execute use case
	input := dashboard.GetPeriodTransactionsInput{
		UserID:     userID,
		StartDate:  startDate,
		EndDate:    endDate,
		CategoryID: categoryID,
		Limit:      limit,
		Offset:     offset,
	}

	output, err := c.getPeriodTransactionsUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleDashboardError(ctx, err)
		return
	}

	// Transform to response DTO
	response := dto.ToPeriodTransactionsResponse(output)
	ctx.JSON(http.StatusOK, response)
}

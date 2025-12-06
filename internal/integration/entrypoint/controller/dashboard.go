// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/finance-tracker/backend/internal/application/usecase/dashboard"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// DashboardController handles dashboard endpoints.
type DashboardController struct {
	getCategoryTrendsUseCase *dashboard.GetCategoryTrendsUseCase
}

// NewDashboardController creates a new dashboard controller instance.
func NewDashboardController(
	getCategoryTrendsUseCase *dashboard.GetCategoryTrendsUseCase,
) *DashboardController {
	return &DashboardController{
		getCategoryTrendsUseCase: getCategoryTrendsUseCase,
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

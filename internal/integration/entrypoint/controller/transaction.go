// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/finance-tracker/backend/internal/application/usecase/transaction"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// TransactionController handles transaction endpoints.
type TransactionController struct {
	listUseCase          *transaction.ListTransactionsUseCase
	createUseCase        *transaction.CreateTransactionUseCase
	updateUseCase        *transaction.UpdateTransactionUseCase
	deleteUseCase        *transaction.DeleteTransactionUseCase
	bulkDeleteUseCase    *transaction.BulkDeleteTransactionsUseCase
	bulkCategorizeUseCase *transaction.BulkCategorizeTransactionsUseCase
}

// NewTransactionController creates a new transaction controller instance.
func NewTransactionController(
	listUseCase *transaction.ListTransactionsUseCase,
	createUseCase *transaction.CreateTransactionUseCase,
	updateUseCase *transaction.UpdateTransactionUseCase,
	deleteUseCase *transaction.DeleteTransactionUseCase,
	bulkDeleteUseCase *transaction.BulkDeleteTransactionsUseCase,
	bulkCategorizeUseCase *transaction.BulkCategorizeTransactionsUseCase,
) *TransactionController {
	return &TransactionController{
		listUseCase:          listUseCase,
		createUseCase:        createUseCase,
		updateUseCase:        updateUseCase,
		deleteUseCase:        deleteUseCase,
		bulkDeleteUseCase:    bulkDeleteUseCase,
		bulkCategorizeUseCase: bulkCategorizeUseCase,
	}
}

// List handles GET /transactions requests.
func (c *TransactionController) List(ctx *gin.Context) {
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
	input := transaction.ListTransactionsInput{
		UserID: userID,
	}

	// Parse date filters
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

	// Parse category IDs filter
	if categoryIDsStr := ctx.Query("categoryIds"); categoryIDsStr != "" {
		ids := strings.Split(categoryIDsStr, ",")
		for _, idStr := range ids {
			if id, err := uuid.Parse(strings.TrimSpace(idStr)); err == nil {
				input.CategoryIDs = append(input.CategoryIDs, id)
			}
		}
	}

	// Parse type filter
	if typeStr := ctx.Query("type"); typeStr != "" {
		txnType := entity.TransactionType(typeStr)
		input.Type = &txnType
	}

	// Parse search filter
	input.Search = ctx.Query("search")

	// Parse groupByDate flag
	if groupByDateStr := ctx.Query("groupByDate"); groupByDateStr == "true" {
		input.GroupByDate = true
	}

	// Parse pagination
	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			input.Page = page
		}
	}
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			input.Limit = limit
		}
	}

	// Execute use case
	output, err := c.listUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve transactions",
		})
		return
	}

	// Build response
	response := dto.ToTransactionListResponse(output)
	ctx.JSON(http.StatusOK, response)
}

// Create handles POST /transactions requests.
func (c *TransactionController) Create(ctx *gin.Context) {
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
	var req dto.CreateTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
			Code:  string(domainerror.ErrCodeMissingTransactionFields),
		})
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid date format. Use YYYY-MM-DD",
			Code:  string(domainerror.ErrCodeInvalidTransactionDate),
		})
		return
	}

	// Parse category ID if provided
	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid category ID format",
			})
			return
		}
		categoryID = &id
	}

	// Build input
	input := transaction.CreateTransactionInput{
		UserID:              userID,
		Date:                date,
		Description:         req.Description,
		Amount:              decimal.NewFromFloat(req.Amount),
		Type:                entity.TransactionType(req.Type),
		CategoryID:          categoryID,
		Notes:               req.Notes,
		IsRecurring:         req.IsRecurring,
		BillingCycle:        req.BillingCycle,
		IsCreditCardPayment: req.IsCreditCardPayment,
	}

	// Execute use case
	output, err := c.createUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleTransactionError(ctx, err)
		return
	}

	// Build response
	response := dto.ToTransactionResponse(output.Transaction)
	ctx.JSON(http.StatusCreated, response)
}

// Update handles PATCH /transactions/:id requests.
func (c *TransactionController) Update(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse transaction ID from URL
	transactionIDStr := ctx.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid transaction ID format",
		})
		return
	}

	// Parse request body
	var req dto.UpdateTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Build input
	input := transaction.UpdateTransactionInput{
		TransactionID: transactionID,
		UserID:        userID,
		ClearCategory: req.ClearCategory,
	}

	// Parse optional fields
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid date format. Use YYYY-MM-DD",
				Code:  string(domainerror.ErrCodeInvalidTransactionDate),
			})
			return
		}
		input.Date = &date
	}

	if req.Description != nil {
		input.Description = req.Description
	}

	if req.Amount != nil {
		amount := decimal.NewFromFloat(*req.Amount)
		input.Amount = &amount
	}

	if req.Type != nil {
		txnType := entity.TransactionType(*req.Type)
		input.Type = &txnType
	}

	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid category ID format",
			})
			return
		}
		input.CategoryID = &id
	}

	if req.Notes != nil {
		input.Notes = req.Notes
	}

	if req.IsRecurring != nil {
		input.IsRecurring = req.IsRecurring
	}

	// Execute use case
	output, err := c.updateUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleTransactionError(ctx, err)
		return
	}

	// Build response
	response := dto.ToTransactionResponse(output.Transaction)
	ctx.JSON(http.StatusOK, response)
}

// Delete handles DELETE /transactions/:id requests.
func (c *TransactionController) Delete(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse transaction ID from URL
	transactionIDStr := ctx.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid transaction ID format",
		})
		return
	}

	// Build input
	input := transaction.DeleteTransactionInput{
		TransactionID: transactionID,
		UserID:        userID,
	}

	// Execute use case
	_, err = c.deleteUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleTransactionError(ctx, err)
		return
	}

	// Return no content on success
	ctx.Status(http.StatusNoContent)
}

// BulkDelete handles POST /transactions/bulk-delete requests.
func (c *TransactionController) BulkDelete(ctx *gin.Context) {
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
	var req dto.BulkDeleteTransactionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Parse transaction IDs
	transactionIDs := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid transaction ID format: " + idStr,
			})
			return
		}
		transactionIDs = append(transactionIDs, id)
	}

	// Build input
	input := transaction.BulkDeleteTransactionsInput{
		TransactionIDs: transactionIDs,
		UserID:         userID,
	}

	// Execute use case
	output, err := c.bulkDeleteUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleTransactionError(ctx, err)
		return
	}

	// Build response
	response := dto.BulkDeleteTransactionsResponse{
		DeletedCount: output.DeletedCount,
	}
	ctx.JSON(http.StatusOK, response)
}

// BulkCategorize handles POST /transactions/bulk-categorize requests.
func (c *TransactionController) BulkCategorize(ctx *gin.Context) {
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
	var req dto.BulkCategorizeTransactionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Parse transaction IDs
	transactionIDs := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid transaction ID format: " + idStr,
			})
			return
		}
		transactionIDs = append(transactionIDs, id)
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
	input := transaction.BulkCategorizeTransactionsInput{
		TransactionIDs: transactionIDs,
		CategoryID:     categoryID,
		UserID:         userID,
	}

	// Execute use case
	output, err := c.bulkCategorizeUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleTransactionError(ctx, err)
		return
	}

	// Build response
	response := dto.BulkCategorizeTransactionsResponse{
		UpdatedCount: output.UpdatedCount,
	}
	ctx.JSON(http.StatusOK, response)
}

// handleTransactionError handles transaction errors and returns appropriate HTTP responses.
func (c *TransactionController) handleTransactionError(ctx *gin.Context, err error) {
	var txnErr *domainerror.TransactionError
	if errors.As(err, &txnErr) {
		statusCode := c.getStatusCodeForTransactionError(txnErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: txnErr.Message,
			Code:  string(txnErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForTransactionError maps transaction error codes to HTTP status codes.
func (c *TransactionController) getStatusCodeForTransactionError(code domainerror.TransactionErrorCode) int {
	switch code {
	case domainerror.ErrCodeTransactionNotFound,
		domainerror.ErrCodeTxnCategoryNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeNotAuthorizedTransaction,
		domainerror.ErrCodeTxnCategoryNotOwned:
		return http.StatusForbidden
	case domainerror.ErrCodeInvalidTransactionType,
		domainerror.ErrCodeInvalidTransactionDate,
		domainerror.ErrCodeInvalidTransactionAmount,
		domainerror.ErrCodeDescriptionTooLong,
		domainerror.ErrCodeNotesTooLong,
		domainerror.ErrCodeMissingTransactionFields,
		domainerror.ErrCodeEmptyTransactionIDs:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

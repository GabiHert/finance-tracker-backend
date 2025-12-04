// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	creditcard "github.com/finance-tracker/backend/internal/application/usecase/credit_card"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// CreditCardController handles credit card import endpoints.
type CreditCardController struct {
	previewImportUseCase     *creditcard.PreviewImportUseCase
	importTransactionsUseCase *creditcard.ImportTransactionsUseCase
	collapseExpansionUseCase  *creditcard.CollapseExpansionUseCase
	getStatusUseCase          *creditcard.GetStatusUseCase
}

// NewCreditCardController creates a new credit card controller instance.
func NewCreditCardController(
	previewImportUseCase *creditcard.PreviewImportUseCase,
	importTransactionsUseCase *creditcard.ImportTransactionsUseCase,
	collapseExpansionUseCase *creditcard.CollapseExpansionUseCase,
	getStatusUseCase *creditcard.GetStatusUseCase,
) *CreditCardController {
	return &CreditCardController{
		previewImportUseCase:     previewImportUseCase,
		importTransactionsUseCase: importTransactionsUseCase,
		collapseExpansionUseCase:  collapseExpansionUseCase,
		getStatusUseCase:          getStatusUseCase,
	}
}

// Preview handles POST /transactions/credit-card/preview requests.
func (c *CreditCardController) Preview(ctx *gin.Context) {
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
	var req dto.ImportPreviewRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Convert DTOs to use case input
	transactions := make([]creditcard.CCTransactionInput, len(req.Transactions))
	for i, txnDTO := range req.Transactions {
		date, err := time.Parse("2006-01-02", txnDTO.Date)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid date format for transaction at index " + string(rune(i)),
				Code:  string(domainerror.ErrCodeInvalidTransactionDate),
			})
			return
		}

		transactions[i] = creditcard.CCTransactionInput{
			Date:               date,
			Description:        txnDTO.Description,
			Amount:             decimal.NewFromFloat(txnDTO.Amount),
			InstallmentCurrent: txnDTO.InstallmentCurrent,
			InstallmentTotal:   txnDTO.InstallmentTotal,
		}
	}

	// Build input
	input := creditcard.PreviewImportInput{
		UserID:       userID,
		BillingCycle: req.BillingCycle,
		Transactions: transactions,
	}

	// Execute use case
	output, err := c.previewImportUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCreditCardError(ctx, err)
		return
	}

	// Build response
	potentialMatches := make([]dto.BillMatchDTO, len(output.PotentialMatches))
	for i, match := range output.PotentialMatches {
		potentialMatches[i] = dto.ToBillMatchDTO(
			match.BillPaymentID,
			match.BillPaymentDate,
			match.BillPaymentAmount,
			match.BillDescription,
			match.CCPaymentDate,
			match.CCPaymentAmount,
			match.MatchScore,
		)
	}

	transactionsToImport := make([]dto.CreditCardTransactionDTO, len(output.TransactionsToImport))
	for i, txn := range output.TransactionsToImport {
		transactionsToImport[i] = dto.CreditCardTransactionDTO{
			Date:               txn.Date.Format("2006-01-02"),
			Description:        txn.Description,
			Amount:             txn.Amount.InexactFloat64(),
			InstallmentCurrent: txn.InstallmentCurrent,
			InstallmentTotal:   txn.InstallmentTotal,
		}
	}

	response := dto.ImportPreviewResponseDTO{
		BillingCycle:          output.BillingCycle,
		TotalTransactions:     output.TotalTransactions,
		TotalAmount:           output.TotalAmount.String(),
		PotentialMatches:      potentialMatches,
		TransactionsToImport:  transactionsToImport,
		PaymentReceivedAmount: output.PaymentReceivedAmount.String(),
		HasExistingImport:     output.HasExistingImport,
	}

	ctx.JSON(http.StatusOK, response)
}

// Import handles POST /transactions/credit-card/import requests.
func (c *CreditCardController) Import(ctx *gin.Context) {
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
	var req dto.ImportRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Parse bill payment ID
	billPaymentID, err := uuid.Parse(req.BillPaymentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid bill payment ID format",
		})
		return
	}

	// Convert DTOs to use case input
	transactions := make([]creditcard.CCTransactionInput, len(req.Transactions))
	for i, txnDTO := range req.Transactions {
		date, err := time.Parse("2006-01-02", txnDTO.Date)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid date format for transaction",
				Code:  string(domainerror.ErrCodeInvalidTransactionDate),
			})
			return
		}

		transactions[i] = creditcard.CCTransactionInput{
			Date:               date,
			Description:        txnDTO.Description,
			Amount:             decimal.NewFromFloat(txnDTO.Amount),
			InstallmentCurrent: txnDTO.InstallmentCurrent,
			InstallmentTotal:   txnDTO.InstallmentTotal,
		}
	}

	// Build input
	input := creditcard.ImportTransactionsInput{
		UserID:            userID,
		BillingCycle:      req.BillingCycle,
		BillPaymentID:     billPaymentID,
		Transactions:      transactions,
		ApplyAutoCategory: req.ApplyAutoCategory,
	}

	// Execute use case
	output, err := c.importTransactionsUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCreditCardError(ctx, err)
		return
	}

	// Build response
	transactionSummaries := make([]dto.ImportedTransactionSummary, len(output.Transactions))
	for i, txn := range output.Transactions {
		summary := dto.ImportedTransactionSummary{
			ID:          txn.ID.String(),
			Date:        txn.Date.Format("2006-01-02"),
			Description: txn.Description,
			Amount:      txn.Amount.String(),
		}
		if txn.CategoryID != nil {
			catIDStr := txn.CategoryID.String()
			summary.CategoryID = &catIDStr
		}
		transactionSummaries[i] = summary
	}

	response := dto.ImportResultDTO{
		ImportedCount:      output.ImportedCount,
		CategorizedCount:   output.CategorizedCount,
		BillPaymentID:      output.BillPaymentID.String(),
		BillingCycle:       output.BillingCycle,
		OriginalBillAmount: output.OriginalBillAmount.String(),
		ImportedAt:         output.ImportedAt,
		Transactions:       transactionSummaries,
	}

	ctx.JSON(http.StatusCreated, response)
}

// Collapse handles POST /transactions/credit-card/collapse requests.
func (c *CreditCardController) Collapse(ctx *gin.Context) {
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
	var req dto.CollapseRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Parse bill payment ID
	billPaymentID, err := uuid.Parse(req.BillPaymentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid bill payment ID format",
		})
		return
	}

	// Build input
	input := creditcard.CollapseExpansionInput{
		UserID:        userID,
		BillPaymentID: billPaymentID,
	}

	// Execute use case
	output, err := c.collapseExpansionUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCreditCardError(ctx, err)
		return
	}

	// Build response
	response := dto.CollapseResultDTO{
		BillPaymentID:       output.BillPaymentID.String(),
		RestoredAmount:      output.RestoredAmount.String(),
		DeletedTransactions: output.DeletedTransactions,
		CollapsedAt:         output.CollapsedAt,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetStatus handles GET /transactions/credit-card/status requests.
func (c *CreditCardController) GetStatus(ctx *gin.Context) {
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
	billingCycle := ctx.Query("billing_cycle")

	// Build input
	input := creditcard.GetStatusInput{
		UserID:       userID,
		BillingCycle: billingCycle,
	}

	// Execute use case
	output, err := c.getStatusUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleCreditCardError(ctx, err)
		return
	}

	// Build response
	transactionSummaries := make([]dto.CCTransactionSummaryDTO, len(output.LinkedTransactions))
	for i, txn := range output.LinkedTransactions {
		transactionSummaries[i] = dto.ToCreditCardTransactionDTO(
			txn.ID,
			txn.Date,
			txn.Description,
			txn.Amount,
			txn.CategoryID,
			txn.IsHidden,
		)
	}

	response := dto.CreditCardStatusDTO{
		BillingCycle:        output.BillingCycle,
		IsExpanded:          output.IsExpanded,
		LinkedTransactions:  len(output.LinkedTransactions),
		TransactionsSummary: transactionSummaries,
		ExpandedAt:          output.ExpandedAt,
	}

	if output.BillPaymentID != nil {
		billIDStr := output.BillPaymentID.String()
		response.BillPaymentID = &billIDStr
	}

	if output.BillPaymentDate != nil {
		billDateStr := output.BillPaymentDate.Format("2006-01-02")
		response.BillPaymentDate = &billDateStr
	}

	if output.OriginalAmount != nil {
		originalAmtStr := output.OriginalAmount.String()
		response.OriginalAmount = &originalAmtStr
	}

	if output.CurrentAmount != nil {
		currentAmtStr := output.CurrentAmount.String()
		response.CurrentAmount = &currentAmtStr
	}

	ctx.JSON(http.StatusOK, response)
}

// handleCreditCardError handles credit card errors and returns appropriate HTTP responses.
func (c *CreditCardController) handleCreditCardError(ctx *gin.Context, err error) {
	var txnErr *domainerror.TransactionError
	if errors.As(err, &txnErr) {
		statusCode := c.getStatusCodeForCreditCardError(txnErr.Code)
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

// getStatusCodeForCreditCardError maps credit card error codes to HTTP status codes.
func (c *CreditCardController) getStatusCodeForCreditCardError(code domainerror.TransactionErrorCode) int {
	switch code {
	case domainerror.ErrCodeBillPaymentNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeBillPaymentNotOwned:
		return http.StatusForbidden
	case domainerror.ErrCodeInvalidBillingCycle,
		domainerror.ErrCodeEmptyCCTransactions,
		domainerror.ErrCodeBillNotExpanded,
		domainerror.ErrCodeBillAlreadyExpanded,
		domainerror.ErrCodeNoPotentialMatches:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

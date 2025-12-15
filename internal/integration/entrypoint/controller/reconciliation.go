// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/usecase/reconciliation"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// ReconciliationController handles reconciliation endpoints.
type ReconciliationController struct {
	getPendingUseCase             *reconciliation.GetPendingUseCase
	getLinkedUseCase              *reconciliation.GetLinkedUseCase
	getSummaryUseCase             *reconciliation.GetSummaryUseCase
	manualLinkUseCase             *reconciliation.ManualLinkUseCase
	unlinkUseCase                 *reconciliation.UnlinkUseCase
	triggerReconciliationUseCase  *reconciliation.TriggerReconciliationUseCase
}

// NewReconciliationController creates a new reconciliation controller instance.
func NewReconciliationController(
	getPendingUseCase *reconciliation.GetPendingUseCase,
	getLinkedUseCase *reconciliation.GetLinkedUseCase,
	getSummaryUseCase *reconciliation.GetSummaryUseCase,
	manualLinkUseCase *reconciliation.ManualLinkUseCase,
	unlinkUseCase *reconciliation.UnlinkUseCase,
	triggerReconciliationUseCase *reconciliation.TriggerReconciliationUseCase,
) *ReconciliationController {
	return &ReconciliationController{
		getPendingUseCase:            getPendingUseCase,
		getLinkedUseCase:             getLinkedUseCase,
		getSummaryUseCase:            getSummaryUseCase,
		manualLinkUseCase:            manualLinkUseCase,
		unlinkUseCase:                unlinkUseCase,
		triggerReconciliationUseCase: triggerReconciliationUseCase,
	}
}

// GetPending handles GET /reconciliation/pending requests.
func (c *ReconciliationController) GetPending(ctx *gin.Context) {
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
	limit := 12
	offset := 0
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build input
	input := reconciliation.GetPendingInput{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	// Execute use case
	output, err := c.getPendingUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		// Log error for debugging
		println("[ERROR] GetPending use case failed:", err.Error())
		c.handleReconciliationError(ctx, err)
		return
	}

	// Build response
	pendingCycles := make([]dto.PendingCycleDTO, len(output.PendingCycles))
	for i, cycle := range output.PendingCycles {
		potentialBills := make([]dto.PotentialBillDTO, len(cycle.PotentialBills))
		for j, bill := range cycle.PotentialBills {
			potentialBills[j] = dto.ToPotentialBillDTO(
				bill.BillID,
				bill.BillDate,
				bill.BillDescription,
				bill.BillAmount,
				bill.CategoryName,
				bill.Confidence,
				bill.AmountDifference,
				bill.AmountDifferencePercent,
				bill.Score,
			)
		}
		pendingCycles[i] = dto.ToPendingCycleDTO(
			cycle.BillingCycle,
			cycle.DisplayName,
			cycle.TransactionCount,
			cycle.TotalAmount,
			cycle.OldestDate,
			cycle.NewestDate,
			potentialBills,
		)
	}

	response := dto.GetPendingResponseDTO{
		PendingCycles: pendingCycles,
		Summary: dto.ReconciliationSummaryDTO{
			TotalPending:  output.Summary.TotalPending,
			TotalLinked:   output.Summary.TotalLinked,
			MonthsCovered: output.Summary.MonthsCovered,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// GetLinked handles GET /reconciliation/linked requests.
func (c *ReconciliationController) GetLinked(ctx *gin.Context) {
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
	limit := 12
	offset := 0
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build input
	input := reconciliation.GetLinkedInput{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	// Execute use case
	output, err := c.getLinkedUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleReconciliationError(ctx, err)
		return
	}

	// Build response
	linkedCycles := make([]dto.LinkedCycleDTO, len(output.LinkedCycles))
	for i, cycle := range output.LinkedCycles {
		linkedCycles[i] = dto.ToLinkedCycleDTO(
			cycle.BillingCycle,
			cycle.DisplayName,
			cycle.TransactionCount,
			cycle.TotalAmount,
			cycle.Bill.ID,
			cycle.Bill.Date,
			cycle.Bill.Description,
			cycle.Bill.OriginalAmount,
			cycle.Bill.CategoryName,
			cycle.AmountDifference,
			cycle.HasMismatch,
		)
	}

	response := dto.GetLinkedResponseDTO{
		LinkedCycles: linkedCycles,
		Summary: dto.ReconciliationSummaryDTO{
			TotalPending:  output.Summary.TotalPending,
			TotalLinked:   output.Summary.TotalLinked,
			MonthsCovered: output.Summary.MonthsCovered,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// GetSummary handles GET /reconciliation/summary requests.
func (c *ReconciliationController) GetSummary(ctx *gin.Context) {
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
	input := reconciliation.GetSummaryInput{
		UserID: userID,
	}

	// Execute use case
	output, err := c.getSummaryUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleReconciliationError(ctx, err)
		return
	}

	// Build response
	response := dto.GetSummaryResponseDTO{
		TotalPending:  output.TotalPending,
		TotalLinked:   output.TotalLinked,
		MonthsCovered: output.MonthsCovered,
	}

	ctx.JSON(http.StatusOK, response)
}

// ManualLink handles POST /reconciliation/link requests.
func (c *ReconciliationController) ManualLink(ctx *gin.Context) {
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
	var req dto.ManualLinkRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Parse bill payment ID
	billID, err := uuid.Parse(req.BillPaymentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid bill payment ID format",
		})
		return
	}

	// Build input
	input := reconciliation.ManualLinkInput{
		UserID:       userID,
		BillingCycle: req.BillingCycle,
		BillID:       billID,
		Force:        req.Force,
	}

	// Execute use case
	output, err := c.manualLinkUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleReconciliationError(ctx, err)
		return
	}

	// Build response
	response := dto.ManualLinkResponseDTO{
		BillingCycle:       output.BillingCycle,
		BillPaymentID:      output.BillID.String(),
		TransactionsLinked: output.TransactionsLinked,
		AmountDifference:   output.AmountDifference.String(),
		HasMismatch:        output.HasMismatch,
	}

	ctx.JSON(http.StatusOK, response)
}

// Unlink handles POST /reconciliation/unlink requests.
func (c *ReconciliationController) Unlink(ctx *gin.Context) {
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
	var req dto.UnlinkRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Build input
	input := reconciliation.UnlinkInput{
		UserID:       userID,
		BillingCycle: req.BillingCycle,
	}

	// Execute use case
	output, err := c.unlinkUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleReconciliationError(ctx, err)
		return
	}

	// Build response
	response := dto.UnlinkResponseDTO{
		BillingCycle: output.BillingCycle,
		Success:      output.Success,
	}

	ctx.JSON(http.StatusOK, response)
}

// TriggerReconciliation handles POST /reconciliation/trigger requests.
func (c *ReconciliationController) TriggerReconciliation(ctx *gin.Context) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	// Parse request body (optional)
	var req dto.TriggerReconciliationRequestDTO
	_ = ctx.ShouldBindJSON(&req) // Ignore error as body is optional

	// Build input
	input := reconciliation.TriggerReconciliationInput{
		UserID: userID,
	}
	if req.BillingCycle != "" {
		input.BillingCycle = &req.BillingCycle
	}

	// Execute use case
	output, err := c.triggerReconciliationUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleReconciliationError(ctx, err)
		return
	}

	// Build response
	autoLinked := make([]dto.AutoLinkedCycleDTO, len(output.AutoLinked))
	for i, cycle := range output.AutoLinked {
		autoLinked[i] = dto.AutoLinkedCycleDTO{
			BillingCycle:     cycle.BillingCycle,
			BillID:           cycle.BillID.String(),
			BillDescription:  cycle.BillDescription,
			TransactionCount: cycle.TransactionCount,
			Confidence:       string(cycle.Confidence),
			AmountDifference: cycle.AmountDifference.String(),
		}
	}

	requiresSelection := make([]dto.PendingWithMatchesDTO, len(output.RequiresSelection))
	for i, cycle := range output.RequiresSelection {
		potentialBills := make([]dto.PotentialBillDTO, len(cycle.PotentialBills))
		for j, bill := range cycle.PotentialBills {
			potentialBills[j] = dto.ToPotentialBillDTO(
				bill.BillID,
				bill.BillDate,
				bill.BillDescription,
				bill.BillAmount,
				bill.CategoryName,
				bill.Confidence,
				bill.AmountDifference,
				bill.AmountDifferencePercent,
				bill.Score,
			)
		}
		requiresSelection[i] = dto.PendingWithMatchesDTO{
			BillingCycle:   cycle.BillingCycle,
			PotentialBills: potentialBills,
		}
	}

	noMatch := make([]dto.NoMatchCycleDTO, len(output.NoMatch))
	for i, cycle := range output.NoMatch {
		noMatch[i] = dto.NoMatchCycleDTO{
			BillingCycle:     cycle.BillingCycle,
			TransactionCount: cycle.TransactionCount,
			TotalAmount:      cycle.TotalAmount.String(),
		}
	}

	response := dto.TriggerReconciliationResponseDTO{
		AutoLinked:        autoLinked,
		RequiresSelection: requiresSelection,
		NoMatch:           noMatch,
		Summary: dto.ReconciliationResultSummaryDTO{
			AutoLinked:        output.Summary.AutoLinked,
			RequiresSelection: output.Summary.RequiresSelection,
			NoMatch:           output.Summary.NoMatch,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// handleReconciliationError handles reconciliation errors and returns appropriate HTTP responses.
func (c *ReconciliationController) handleReconciliationError(ctx *gin.Context, err error) {
	var txnErr *domainerror.TransactionError
	if errors.As(err, &txnErr) {
		statusCode := c.getStatusCodeForReconciliationError(txnErr.Code)
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

// getStatusCodeForReconciliationError maps error codes to HTTP status codes.
func (c *ReconciliationController) getStatusCodeForReconciliationError(code domainerror.TransactionErrorCode) int {
	switch code {
	case domainerror.ErrCodeBillPaymentNotFound,
		domainerror.ErrCodePendingNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeBillPaymentNotOwned:
		return http.StatusForbidden
	case domainerror.ErrCodeInvalidBillingCycle,
		domainerror.ErrCodeAmountMismatch:
		return http.StatusBadRequest
	case domainerror.ErrCodeCycleAlreadyLinked,
		domainerror.ErrCodeBillAlreadyExpanded:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

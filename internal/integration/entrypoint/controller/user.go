// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/finance-tracker/backend/internal/application/usecase/auth"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// UserController handles user management endpoints.
type UserController struct {
	deleteAccountUseCase *auth.DeleteAccountUseCase
}

// NewUserController creates a new user controller instance.
func NewUserController(
	deleteAccountUseCase *auth.DeleteAccountUseCase,
) *UserController {
	return &UserController{
		deleteAccountUseCase: deleteAccountUseCase,
	}
}

// DeleteAccount handles DELETE /users/me requests.
func (c *UserController) DeleteAccount(ctx *gin.Context) {
	// Get user ID from auth context
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Unauthorized",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	var req dto.DeleteAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingFields),
		})
		return
	}

	input := auth.DeleteAccountInput{
		UserID:       userID,
		Password:     req.Password,
		Confirmation: req.Confirmation,
	}

	_, err := c.deleteAccountUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleDeleteAccountError(ctx, err)
		return
	}

	// Return 204 No Content on successful deletion
	ctx.Status(http.StatusNoContent)
}

// handleDeleteAccountError handles delete account errors and returns appropriate HTTP responses.
func (c *UserController) handleDeleteAccountError(ctx *gin.Context, err error) {
	var authErr *domainerror.AuthError
	if errors.As(err, &authErr) {
		statusCode := c.getStatusCodeForDeleteAccountError(authErr.Code)
		ctx.JSON(statusCode, dto.ErrorResponse{
			Error: authErr.Message,
			Code:  string(authErr.Code),
		})
		return
	}

	// Generic server error
	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: "An internal error occurred",
	})
}

// getStatusCodeForDeleteAccountError maps auth error codes to HTTP status codes for delete account.
func (c *UserController) getStatusCodeForDeleteAccountError(code domainerror.AuthErrorCode) int {
	switch code {
	case domainerror.ErrCodeInvalidCredentials:
		return http.StatusUnauthorized
	case domainerror.ErrCodeUserNotFound:
		return http.StatusNotFound
	case domainerror.ErrCodeInvalidConfirmation,
		domainerror.ErrCodeMissingFields:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

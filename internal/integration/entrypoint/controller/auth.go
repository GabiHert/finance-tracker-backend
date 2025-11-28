// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/finance-tracker/backend/internal/application/usecase/auth"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
)

// AuthController handles authentication endpoints.
type AuthController struct {
	registerUseCase       *auth.RegisterUserUseCase
	loginUseCase          *auth.LoginUserUseCase
	refreshTokenUseCase   *auth.RefreshTokenUseCase
	logoutUseCase         *auth.LogoutUserUseCase
	forgotPasswordUseCase *auth.ForgotPasswordUseCase
	resetPasswordUseCase  *auth.ResetPasswordUseCase
}

// NewAuthController creates a new auth controller instance.
func NewAuthController(
	registerUseCase *auth.RegisterUserUseCase,
	loginUseCase *auth.LoginUserUseCase,
	refreshTokenUseCase *auth.RefreshTokenUseCase,
	logoutUseCase *auth.LogoutUserUseCase,
	forgotPasswordUseCase *auth.ForgotPasswordUseCase,
	resetPasswordUseCase *auth.ResetPasswordUseCase,
) *AuthController {
	return &AuthController{
		registerUseCase:       registerUseCase,
		loginUseCase:          loginUseCase,
		refreshTokenUseCase:   refreshTokenUseCase,
		logoutUseCase:         logoutUseCase,
		forgotPasswordUseCase: forgotPasswordUseCase,
		resetPasswordUseCase:  resetPasswordUseCase,
	}
}

// Register handles POST /auth/register requests.
func (c *AuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingFields),
		})
		return
	}

	input := auth.RegisterUserInput{
		Email:         req.Email,
		Name:          req.Name,
		Password:      req.Password,
		TermsAccepted: req.TermsAccepted,
	}

	output, err := c.registerUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.AuthResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
		User:         dto.ToUserResponse(output.User),
	})
}

// Login handles POST /auth/login requests.
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingFields),
		})
		return
	}

	input := auth.LoginUserInput{
		Email:      req.Email,
		Password:   req.Password,
		RememberMe: req.RememberMe,
	}

	output, err := c.loginUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
		User:         dto.ToUserResponse(output.User),
	})
}

// RefreshToken handles POST /auth/refresh requests.
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingToken),
		})
		return
	}

	input := auth.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	}

	output, err := c.refreshTokenUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	})
}

// Logout handles POST /auth/logout requests.
func (c *AuthController) Logout(ctx *gin.Context) {
	var req dto.LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Even with invalid body, return success for logout
		ctx.JSON(http.StatusOK, dto.MessageResponse{
			Message: "Successfully logged out",
		})
		return
	}

	input := auth.LogoutUserInput{
		RefreshToken: req.RefreshToken,
	}

	output, _ := c.logoutUseCase.Execute(ctx.Request.Context(), input)

	ctx.JSON(http.StatusOK, dto.MessageResponse{
		Message: output.Message,
	})
}

// ForgotPassword handles POST /auth/forgot-password requests.
func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeInvalidEmail),
		})
		return
	}

	input := auth.ForgotPasswordInput{
		Email: req.Email,
	}

	output, err := c.forgotPasswordUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.MessageResponse{
		Message: output.Message,
	})
}

// ResetPassword handles POST /auth/reset-password requests.
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request body",
			Code:  string(domainerror.ErrCodeMissingFields),
		})
		return
	}

	input := auth.ResetPasswordInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}

	output, err := c.resetPasswordUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.MessageResponse{
		Message: output.Message,
	})
}

// handleAuthError handles authentication errors and returns appropriate HTTP responses.
func (c *AuthController) handleAuthError(ctx *gin.Context, err error) {
	var authErr *domainerror.AuthError
	if errors.As(err, &authErr) {
		statusCode := c.getStatusCodeForAuthError(authErr.Code)
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

// getStatusCodeForAuthError maps auth error codes to HTTP status codes.
func (c *AuthController) getStatusCodeForAuthError(code domainerror.AuthErrorCode) int {
	switch code {
	case domainerror.ErrCodeEmailExists:
		return http.StatusConflict
	case domainerror.ErrCodeTermsNotAccepted,
		domainerror.ErrCodeWeakPassword,
		domainerror.ErrCodeInvalidEmail,
		domainerror.ErrCodeMissingFields,
		domainerror.ErrCodeInvalidResetToken,
		domainerror.ErrCodeExpiredResetToken:
		return http.StatusBadRequest
	case domainerror.ErrCodeInvalidCredentials,
		domainerror.ErrCodeUserNotFound,
		domainerror.ErrCodeInvalidToken,
		domainerror.ErrCodeExpiredToken,
		domainerror.ErrCodeMissingToken:
		return http.StatusUnauthorized
	case domainerror.ErrCodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

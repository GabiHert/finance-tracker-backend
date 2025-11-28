// Package auth contains authentication-related use cases.
package auth

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// ForgotPasswordInput represents the input for forgot password request.
type ForgotPasswordInput struct {
	Email string
}

// ForgotPasswordOutput represents the output of forgot password request.
type ForgotPasswordOutput struct {
	Message string
}

// ForgotPasswordUseCase handles forgot password logic.
type ForgotPasswordUseCase struct {
	userRepo          adapter.UserRepository
	resetTokenService adapter.PasswordResetTokenService
}

// NewForgotPasswordUseCase creates a new ForgotPasswordUseCase instance.
func NewForgotPasswordUseCase(
	userRepo adapter.UserRepository,
	resetTokenService adapter.PasswordResetTokenService,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepo:          userRepo,
		resetTokenService: resetTokenService,
	}
}

// Execute performs the forgot password request.
// Always returns success to prevent email enumeration.
func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, input ForgotPasswordInput) (*ForgotPasswordOutput, error) {
	// Validate email format
	if !isValidEmailFormat(input.Email) {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidEmail,
			"invalid email format",
			domainerror.ErrInvalidEmail,
		)
	}

	// Try to find user by email
	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// User not found, but we still return success to prevent enumeration
		slog.Debug("Forgot password requested for non-existent email", "email", input.Email)
		return &ForgotPasswordOutput{
			Message: "If an account with that email exists, we have sent a password reset link",
		}, nil
	}

	// Generate reset token
	_, err = uc.resetTokenService.GenerateResetToken(ctx, user.ID, user.Email)
	if err != nil {
		// Log error but still return success to prevent enumeration
		slog.Error("Failed to generate reset token", "error", err, "userID", user.ID)
		return &ForgotPasswordOutput{
			Message: "If an account with that email exists, we have sent a password reset link",
		}, nil
	}

	// In a real implementation, we would send an email here
	// For now, we just log it
	slog.Info("Password reset token generated", "userID", user.ID, "email", user.Email)

	return &ForgotPasswordOutput{
		Message: "If an account with that email exists, we have sent a password reset link",
	}, nil
}

// isValidEmailFormat validates email format using a simple regex.
func isValidEmailFormat(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

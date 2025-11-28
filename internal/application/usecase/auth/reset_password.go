// Package auth contains authentication-related use cases.
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// ResetPasswordInput represents the input for password reset.
type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

// ResetPasswordOutput represents the output of password reset.
type ResetPasswordOutput struct {
	Message string
}

// ResetPasswordUseCase handles password reset logic.
type ResetPasswordUseCase struct {
	userRepo          adapter.UserRepository
	passwordService   adapter.PasswordService
	resetTokenService adapter.PasswordResetTokenService
}

// NewResetPasswordUseCase creates a new ResetPasswordUseCase instance.
func NewResetPasswordUseCase(
	userRepo adapter.UserRepository,
	passwordService adapter.PasswordService,
	resetTokenService adapter.PasswordResetTokenService,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepo:          userRepo,
		passwordService:   passwordService,
		resetTokenService: resetTokenService,
	}
}

// Execute performs the password reset.
func (uc *ResetPasswordUseCase) Execute(ctx context.Context, input ResetPasswordInput) (*ResetPasswordOutput, error) {
	// Validate reset token
	resetToken, err := uc.resetTokenService.ValidateResetToken(ctx, input.Token)
	if err != nil {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidResetToken,
			"invalid or expired password reset token",
			domainerror.ErrInvalidResetToken,
		)
	}

	// Check if token has expired
	if time.Now().UTC().After(resetToken.ExpiresAt) {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeExpiredResetToken,
			"password reset token has expired",
			domainerror.ErrInvalidResetToken,
		)
	}

	// Validate new password strength
	if err := uc.passwordService.ValidatePasswordStrength(input.NewPassword); err != nil {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeWeakPassword,
			"password does not meet minimum requirements",
			domainerror.ErrWeakPassword,
		)
	}

	// Find user
	user, err := uc.userRepo.FindByID(ctx, resetToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Hash new password
	passwordHash, err := uc.passwordService.HashPassword(input.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now().UTC()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user password: %w", err)
	}

	// Invalidate the reset token
	if err := uc.resetTokenService.InvalidateResetToken(ctx, input.Token); err != nil {
		// Log but don't fail - password was already reset
		fmt.Printf("failed to invalidate reset token: %v\n", err)
	}

	return &ResetPasswordOutput{
		Message: "Password has been successfully reset",
	}, nil
}

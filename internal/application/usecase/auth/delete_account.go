// Package auth contains authentication-related use cases.
package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// DeleteAccountInput represents the input for account deletion.
type DeleteAccountInput struct {
	UserID       uuid.UUID
	Password     string
	Confirmation string
}

// DeleteAccountOutput represents the output of account deletion.
type DeleteAccountOutput struct {
	Success bool
}

// DeleteAccountUseCase handles account deletion logic.
type DeleteAccountUseCase struct {
	userRepo        adapter.UserRepository
	passwordService adapter.PasswordService
	tokenService    adapter.TokenService
}

// NewDeleteAccountUseCase creates a new DeleteAccountUseCase instance.
func NewDeleteAccountUseCase(
	userRepo adapter.UserRepository,
	passwordService adapter.PasswordService,
	tokenService adapter.TokenService,
) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{
		userRepo:        userRepo,
		passwordService: passwordService,
		tokenService:    tokenService,
	}
}

// Execute performs the account deletion.
func (uc *DeleteAccountUseCase) Execute(ctx context.Context, input DeleteAccountInput) (*DeleteAccountOutput, error) {
	// Validate confirmation text if provided (frontend may validate in UI)
	if input.Confirmation != "" && input.Confirmation != "DELETE" {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidConfirmation,
			"confirmation must be exactly 'DELETE'",
			nil,
		)
	}

	// Find user by ID
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeUserNotFound,
			"user not found",
			err,
		)
	}

	// Verify password
	if err := uc.passwordService.VerifyPassword(user.PasswordHash, input.Password); err != nil {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidCredentials,
			"invalid password",
			domainerror.ErrInvalidCredentials,
		)
	}

	// Invalidate all refresh tokens for the user
	if err := uc.tokenService.InvalidateAllUserTokens(ctx, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to invalidate user tokens: %w", err)
	}

	// Soft delete the user
	if err := uc.userRepo.Delete(ctx, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return &DeleteAccountOutput{
		Success: true,
	}, nil
}

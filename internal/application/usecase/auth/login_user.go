// Package auth contains authentication-related use cases.
package auth

import (
	"context"
	"fmt"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// LoginUserInput represents the input for user login.
type LoginUserInput struct {
	Email      string
	Password   string
	RememberMe bool
}

// LoginUserOutput represents the output of user login.
type LoginUserOutput struct {
	AccessToken  string
	RefreshToken string
	User         *entity.User
}

// LoginUserUseCase handles user login logic.
type LoginUserUseCase struct {
	userRepo        adapter.UserRepository
	passwordService adapter.PasswordService
	tokenService    adapter.TokenService
}

// NewLoginUserUseCase creates a new LoginUserUseCase instance.
func NewLoginUserUseCase(
	userRepo adapter.UserRepository,
	passwordService adapter.PasswordService,
	tokenService adapter.TokenService,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:        userRepo,
		passwordService: passwordService,
		tokenService:    tokenService,
	}
}

// Execute performs the user login.
func (uc *LoginUserUseCase) Execute(ctx context.Context, input LoginUserInput) (*LoginUserOutput, error) {
	// Find user by email
	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Return generic error to prevent email enumeration
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidCredentials,
			"invalid email or password",
			domainerror.ErrInvalidCredentials,
		)
	}

	// Verify password
	if err := uc.passwordService.VerifyPassword(user.PasswordHash, input.Password); err != nil {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidCredentials,
			"invalid email or password",
			domainerror.ErrInvalidCredentials,
		)
	}

	// Generate tokens
	tokenPair, err := uc.tokenService.GenerateTokenPair(ctx, user.ID, user.Email, input.RememberMe)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &LoginUserOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         user,
	}, nil
}

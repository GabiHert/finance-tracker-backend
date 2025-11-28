// Package auth contains authentication-related use cases.
package auth

import (
	"context"

	"github.com/finance-tracker/backend/internal/application/adapter"
)

// LogoutUserInput represents the input for user logout.
type LogoutUserInput struct {
	RefreshToken string
}

// LogoutUserOutput represents the output of user logout.
type LogoutUserOutput struct {
	Message string
}

// LogoutUserUseCase handles user logout logic.
type LogoutUserUseCase struct {
	tokenService adapter.TokenService
}

// NewLogoutUserUseCase creates a new LogoutUserUseCase instance.
func NewLogoutUserUseCase(tokenService adapter.TokenService) *LogoutUserUseCase {
	return &LogoutUserUseCase{
		tokenService: tokenService,
	}
}

// Execute performs the user logout by invalidating the refresh token.
func (uc *LogoutUserUseCase) Execute(ctx context.Context, input LogoutUserInput) (*LogoutUserOutput, error) {
	// Invalidate refresh token (ignore errors as the token might already be invalid)
	_ = uc.tokenService.InvalidateRefreshToken(ctx, input.RefreshToken)

	return &LogoutUserOutput{
		Message: "Successfully logged out",
	}, nil
}

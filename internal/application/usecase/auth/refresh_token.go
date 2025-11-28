// Package auth contains authentication-related use cases.
package auth

import (
	"context"
	"fmt"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// RefreshTokenInput represents the input for token refresh.
type RefreshTokenInput struct {
	RefreshToken string
}

// RefreshTokenOutput represents the output of token refresh.
type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
}

// RefreshTokenUseCase handles token refresh logic.
type RefreshTokenUseCase struct {
	tokenService adapter.TokenService
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase instance.
func NewRefreshTokenUseCase(tokenService adapter.TokenService) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		tokenService: tokenService,
	}
}

// Execute performs the token refresh.
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error) {
	// Validate refresh token
	claims, err := uc.tokenService.ValidateRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidToken,
			"invalid or expired refresh token",
			domainerror.ErrInvalidToken,
		)
	}

	// Check if token is still valid (not invalidated/logged out)
	valid, err := uc.tokenService.IsRefreshTokenValid(ctx, input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to check token validity: %w", err)
	}
	if !valid {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidToken,
			"refresh token has been revoked",
			domainerror.ErrInvalidToken,
		)
	}

	// Invalidate old refresh token
	if err := uc.tokenService.InvalidateRefreshToken(ctx, input.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to invalidate old token: %w", err)
	}

	// Generate new token pair
	tokenPair, err := uc.tokenService.GenerateTokenPair(ctx, claims.UserID, claims.Email, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return &RefreshTokenOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

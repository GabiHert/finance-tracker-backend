// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// TokenClaims represents the claims contained in a JWT token.
type TokenClaims struct {
	UserID    uuid.UUID
	Email     string
	ExpiresAt time.Time
}

// TokenService defines the interface for JWT token operations.
type TokenService interface {
	// GenerateTokenPair generates a new access and refresh token pair.
	GenerateTokenPair(ctx context.Context, userID uuid.UUID, email string, rememberMe bool) (*TokenPair, error)

	// ValidateAccessToken validates an access token and returns its claims.
	ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)

	// ValidateRefreshToken validates a refresh token and returns its claims.
	ValidateRefreshToken(ctx context.Context, token string) (*TokenClaims, error)

	// InvalidateRefreshToken invalidates a refresh token.
	InvalidateRefreshToken(ctx context.Context, token string) error

	// InvalidateAllUserTokens invalidates all refresh tokens for a user.
	InvalidateAllUserTokens(ctx context.Context, userID uuid.UUID) error

	// IsRefreshTokenValid checks if a refresh token is still valid (not invalidated).
	IsRefreshTokenValid(ctx context.Context, token string) (bool, error)
}

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	Token     string
	UserID    uuid.UUID
	Email     string
	ExpiresAt time.Time
}

// PasswordResetTokenService defines the interface for password reset token operations.
type PasswordResetTokenService interface {
	// GenerateResetToken generates a new password reset token.
	GenerateResetToken(ctx context.Context, userID uuid.UUID, email string) (*PasswordResetToken, error)

	// ValidateResetToken validates a password reset token.
	ValidateResetToken(ctx context.Context, token string) (*PasswordResetToken, error)

	// InvalidateResetToken invalidates a password reset token after use.
	InvalidateResetToken(ctx context.Context, token string) error
}

// Package adapters implements adapter interfaces from the application layer.
package adapters

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/integration/persistence"
)

const (
	// Default token durations
	defaultAccessTokenDuration  = 15 * time.Minute
	defaultRefreshTokenDuration = 7 * 24 * time.Hour

	// Extended durations for "remember me" option
	rememberMeAccessTokenDuration  = 7 * 24 * time.Hour
	rememberMeRefreshTokenDuration = 30 * 24 * time.Hour

	// Token types
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

// CustomClaims represents the custom claims for JWT tokens.
type CustomClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// tokenService implements the adapter.TokenService interface.
type tokenService struct {
	secret          []byte
	tokenRepository persistence.TokenRepository
}

// NewTokenService creates a new token service instance.
func NewTokenService(secret string, tokenRepository persistence.TokenRepository) adapter.TokenService {
	return &tokenService{
		secret:          []byte(secret),
		tokenRepository: tokenRepository,
	}
}

// GenerateTokenPair generates a new access and refresh token pair.
func (s *tokenService) GenerateTokenPair(ctx context.Context, userID uuid.UUID, email string, rememberMe bool) (*adapter.TokenPair, error) {
	accessDuration := defaultAccessTokenDuration
	refreshDuration := defaultRefreshTokenDuration

	if rememberMe {
		accessDuration = rememberMeAccessTokenDuration
		refreshDuration = rememberMeRefreshTokenDuration
	}

	// Generate access token
	accessToken, err := s.generateJWT(userID, email, tokenTypeAccess, accessDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.generateJWT(userID, email, tokenTypeRefresh, refreshDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in database
	expiresAt := time.Now().UTC().Add(refreshDuration)
	if err := s.tokenRepository.SaveRefreshToken(ctx, refreshToken, userID, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &adapter.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccessToken validates an access token and returns its claims.
func (s *tokenService) ValidateAccessToken(ctx context.Context, token string) (*adapter.TokenClaims, error) {
	claims, err := s.parseJWT(token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != tokenTypeAccess {
		return nil, fmt.Errorf("invalid token type: expected access token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return &adapter.TokenClaims{
		UserID:    userID,
		Email:     claims.Email,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// ValidateRefreshToken validates a refresh token and returns its claims.
func (s *tokenService) ValidateRefreshToken(ctx context.Context, token string) (*adapter.TokenClaims, error) {
	claims, err := s.parseJWT(token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != tokenTypeRefresh {
		return nil, fmt.Errorf("invalid token type: expected refresh token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return &adapter.TokenClaims{
		UserID:    userID,
		Email:     claims.Email,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// InvalidateRefreshToken invalidates a refresh token.
func (s *tokenService) InvalidateRefreshToken(ctx context.Context, token string) error {
	return s.tokenRepository.InvalidateRefreshToken(ctx, token)
}

// IsRefreshTokenValid checks if a refresh token is still valid (not invalidated).
func (s *tokenService) IsRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	return s.tokenRepository.IsRefreshTokenValid(ctx, token)
}

// generateJWT creates a new JWT token with the given parameters.
func (s *tokenService) generateJWT(userID uuid.UUID, email, tokenType string, duration time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := CustomClaims{
		UserID:    userID.String(),
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "finance-tracker",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// parseJWT parses and validates a JWT token.
func (s *tokenService) parseJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// passwordResetTokenService implements the adapter.PasswordResetTokenService interface.
type passwordResetTokenService struct {
	tokenRepository persistence.TokenRepository
}

// NewPasswordResetTokenService creates a new password reset token service instance.
func NewPasswordResetTokenService(tokenRepository persistence.TokenRepository) adapter.PasswordResetTokenService {
	return &passwordResetTokenService{
		tokenRepository: tokenRepository,
	}
}

// GenerateResetToken generates a new password reset token.
func (s *passwordResetTokenService) GenerateResetToken(ctx context.Context, userID uuid.UUID, email string) (*adapter.PasswordResetToken, error) {
	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Token expires in 1 hour
	expiresAt := time.Now().UTC().Add(1 * time.Hour)

	// Save to database
	if err := s.tokenRepository.SavePasswordResetToken(ctx, token, userID, email, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to save reset token: %w", err)
	}

	return &adapter.PasswordResetToken{
		Token:     token,
		UserID:    userID,
		Email:     email,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateResetToken validates a password reset token.
func (s *passwordResetTokenService) ValidateResetToken(ctx context.Context, token string) (*adapter.PasswordResetToken, error) {
	resetToken, err := s.tokenRepository.GetPasswordResetToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get reset token: %w", err)
	}

	if resetToken == nil {
		return nil, fmt.Errorf("invalid or expired reset token")
	}

	return &adapter.PasswordResetToken{
		Token:     resetToken.Token,
		UserID:    resetToken.UserID,
		Email:     resetToken.Email,
		ExpiresAt: resetToken.ExpiresAt,
	}, nil
}

// InvalidateResetToken invalidates a password reset token after use.
func (s *passwordResetTokenService) InvalidateResetToken(ctx context.Context, token string) error {
	return s.tokenRepository.InvalidatePasswordResetToken(ctx, token)
}

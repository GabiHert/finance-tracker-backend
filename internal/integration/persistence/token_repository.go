// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

// TokenRepository defines the interface for token persistence operations.
type TokenRepository interface {
	// SaveRefreshToken saves a refresh token to the database.
	SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error

	// IsRefreshTokenValid checks if a refresh token is valid (exists and not invalidated).
	IsRefreshTokenValid(ctx context.Context, token string) (bool, error)

	// InvalidateRefreshToken marks a refresh token as invalidated.
	InvalidateRefreshToken(ctx context.Context, token string) error

	// InvalidateAllUserRefreshTokens invalidates all refresh tokens for a user.
	InvalidateAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error

	// SavePasswordResetToken saves a password reset token to the database.
	SavePasswordResetToken(ctx context.Context, token string, userID uuid.UUID, email string, expiresAt time.Time) error

	// GetPasswordResetToken retrieves a password reset token by token string.
	GetPasswordResetToken(ctx context.Context, token string) (*model.PasswordResetTokenModel, error)

	// InvalidatePasswordResetToken marks a password reset token as used.
	InvalidatePasswordResetToken(ctx context.Context, token string) error
}

// tokenRepository implements the TokenRepository interface.
type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository instance.
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{
		db: db,
	}
}

// SaveRefreshToken saves a refresh token to the database.
func (r *tokenRepository) SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	refreshToken := &model.RefreshTokenModel{
		ID:          uuid.New(),
		Token:       token,
		UserID:      userID,
		Invalidated: false,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now().UTC(),
	}
	result := r.db.WithContext(ctx).Create(refreshToken)
	return result.Error
}

// IsRefreshTokenValid checks if a refresh token is valid (exists and not invalidated).
func (r *tokenRepository) IsRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	var refreshToken model.RefreshTokenModel
	result := r.db.WithContext(ctx).
		Where("token = ? AND invalidated = ? AND expires_at > ?", token, false, time.Now().UTC()).
		First(&refreshToken)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

// InvalidateRefreshToken marks a refresh token as invalidated.
func (r *tokenRepository) InvalidateRefreshToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).
		Model(&model.RefreshTokenModel{}).
		Where("token = ?", token).
		Update("invalidated", true)
	return result.Error
}

// InvalidateAllUserRefreshTokens invalidates all refresh tokens for a user.
func (r *tokenRepository) InvalidateAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&model.RefreshTokenModel{}).
		Where("user_id = ?", userID).
		Update("invalidated", true)
	return result.Error
}

// SavePasswordResetToken saves a password reset token to the database.
func (r *tokenRepository) SavePasswordResetToken(ctx context.Context, token string, userID uuid.UUID, email string, expiresAt time.Time) error {
	resetToken := &model.PasswordResetTokenModel{
		ID:        uuid.New(),
		Token:     token,
		UserID:    userID,
		Email:     email,
		Used:      false,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}
	result := r.db.WithContext(ctx).Create(resetToken)
	return result.Error
}

// GetPasswordResetToken retrieves a password reset token by token string.
func (r *tokenRepository) GetPasswordResetToken(ctx context.Context, token string) (*model.PasswordResetTokenModel, error) {
	var resetToken model.PasswordResetTokenModel
	result := r.db.WithContext(ctx).
		Where("token = ? AND used = ?", token, false).
		First(&resetToken)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &resetToken, nil
}

// InvalidatePasswordResetToken marks a password reset token as used.
func (r *tokenRepository) InvalidatePasswordResetToken(ctx context.Context, token string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&model.PasswordResetTokenModel{}).
		Where("token = ?", token).
		Updates(map[string]any{
			"used":    true,
			"used_at": &now,
		})
	return result.Error
}

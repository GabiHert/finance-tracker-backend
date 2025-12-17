// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// AISuggestionRepository defines the interface for AI suggestion persistence operations.
type AISuggestionRepository interface {
	// Create creates a new AI suggestion in the database.
	Create(ctx context.Context, suggestion *entity.AISuggestion) error

	// CreateBatch creates multiple AI suggestions in a single transaction.
	CreateBatch(ctx context.Context, suggestions []*entity.AISuggestion) error

	// GetByID retrieves an AI suggestion by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AISuggestion, error)

	// GetByIDWithDetails retrieves an AI suggestion with all related details.
	GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.AISuggestionWithDetails, error)

	// GetByUserID retrieves all AI suggestions for a given user.
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.AISuggestion, error)

	// GetPendingByUserID retrieves all pending AI suggestions for a given user with details.
	GetPendingByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.AISuggestionWithDetails, error)

	// GetPendingCount retrieves the count of pending AI suggestions for a given user.
	GetPendingCount(ctx context.Context, userID uuid.UUID) (int, error)

	// Update updates an existing AI suggestion in the database.
	Update(ctx context.Context, suggestion *entity.AISuggestion) error

	// DeleteByUserID deletes all AI suggestions for a given user.
	// Returns the number of deleted suggestions.
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// DeleteByID deletes an AI suggestion by its ID.
	DeleteByID(ctx context.Context, id uuid.UUID) error

	// DeletePendingByUserID deletes all pending AI suggestions for a given user.
	// Returns the number of deleted suggestions.
	DeletePendingByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// ExistsPendingByUserID checks if there are any pending suggestions for a user.
	ExistsPendingByUserID(ctx context.Context, userID uuid.UUID) (bool, error)
}

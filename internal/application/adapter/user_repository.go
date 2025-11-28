// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// Create creates a new user in the database.
	Create(ctx context.Context, user *entity.User) error

	// FindByID retrieves a user by their ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// FindByEmail retrieves a user by their email address.
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update updates an existing user in the database.
	Update(ctx context.Context, user *entity.User) error

	// Delete removes a user from the database.
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

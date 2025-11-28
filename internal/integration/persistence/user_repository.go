// Package persistence implements repository interfaces for database operations.
package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

// userRepository implements the adapter.UserRepository interface.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance.
func NewUserRepository(db *gorm.DB) adapter.UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user in the database.
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	userModel := model.FromEntity(user)
	result := r.db.WithContext(ctx).Create(userModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID retrieves a user by their ID.
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var userModel model.UserModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&userModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrUserNotFound
		}
		return nil, result.Error
	}
	return userModel.ToEntity(), nil
}

// FindByEmail retrieves a user by their email address.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var userModel model.UserModel
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&userModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrUserNotFound
		}
		return nil, result.Error
	}
	return userModel.ToEntity(), nil
}

// Update updates an existing user in the database.
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	userModel := model.FromEntity(user)
	result := r.db.WithContext(ctx).Save(userModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete removes a user from the database.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.UserModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.UserModel{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

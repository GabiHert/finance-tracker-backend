// Package adapters implements adapter interfaces from the application layer.
package adapters

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/finance-tracker/backend/internal/application/adapter"
)

const (
	// bcryptCost is the cost factor for bcrypt hashing (12 as per requirements).
	bcryptCost = 12
	// minPasswordLength is the minimum required password length.
	minPasswordLength = 8
)

// passwordService implements the adapter.PasswordService interface.
type passwordService struct{}

// NewPasswordService creates a new password service instance.
func NewPasswordService() adapter.PasswordService {
	return &passwordService{}
}

// HashPassword hashes a plain text password using bcrypt with cost 12.
func (s *passwordService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword compares a plain text password with a hashed password.
func (s *passwordService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePasswordStrength validates if a password meets minimum requirements.
func (s *passwordService) ValidatePasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

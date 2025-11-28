// Package adapter defines interfaces that will be implemented in the integration layer.
package adapter

// PasswordService defines the interface for password hashing and verification.
type PasswordService interface {
	// HashPassword hashes a plain text password using bcrypt.
	HashPassword(password string) (string, error)

	// VerifyPassword compares a plain text password with a hashed password.
	VerifyPassword(hashedPassword, password string) error

	// ValidatePasswordStrength validates if a password meets minimum requirements.
	ValidatePasswordStrength(password string) error
}

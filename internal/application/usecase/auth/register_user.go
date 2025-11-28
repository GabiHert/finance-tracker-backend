// Package auth contains authentication-related use cases.
package auth

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

// RegisterUserInput represents the input for user registration.
type RegisterUserInput struct {
	Email         string
	Name          string
	Password      string
	TermsAccepted bool
}

// RegisterUserOutput represents the output of user registration.
type RegisterUserOutput struct {
	AccessToken  string
	RefreshToken string
	User         *entity.User
}

// RegisterUserUseCase handles user registration logic.
type RegisterUserUseCase struct {
	userRepo        adapter.UserRepository
	passwordService adapter.PasswordService
	tokenService    adapter.TokenService
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase instance.
func NewRegisterUserUseCase(
	userRepo adapter.UserRepository,
	passwordService adapter.PasswordService,
	tokenService adapter.TokenService,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:        userRepo,
		passwordService: passwordService,
		tokenService:    tokenService,
	}
}

// Execute performs the user registration.
func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (*RegisterUserOutput, error) {
	// Validate terms acceptance
	if !input.TermsAccepted {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeTermsNotAccepted,
			"terms of service must be accepted",
			domainerror.ErrTermsNotAccepted,
		)
	}

	// Validate email format
	if !isValidEmail(input.Email) {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeInvalidEmail,
			"invalid email format",
			domainerror.ErrInvalidEmail,
		)
	}

	// Validate password strength
	if err := uc.passwordService.ValidatePasswordStrength(input.Password); err != nil {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeWeakPassword,
			"password does not meet minimum requirements",
			domainerror.ErrWeakPassword,
		)
	}

	// Check if email already exists
	exists, err := uc.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, domainerror.NewAuthError(
			domainerror.ErrCodeEmailExists,
			"email already exists",
			domainerror.ErrEmailAlreadyExists,
		)
	}

	// Hash password
	passwordHash, err := uc.passwordService.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := entity.NewUser(input.Email, input.Name, passwordHash, time.Now().UTC())

	// Save user to database
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenPair, err := uc.tokenService.GenerateTokenPair(ctx, user.ID, user.Email, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &RegisterUserOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         user,
	}, nil
}

// isValidEmail validates email format using a simple regex.
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

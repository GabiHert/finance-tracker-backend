// Package dto defines data transfer objects for API requests and responses.
package dto

import (
	"time"

	"github.com/finance-tracker/backend/internal/domain/entity"
)

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Name          string `json:"name" binding:"required,min=1,max=100"`
	Password      string `json:"password" binding:"required,min=8"`
	TermsAccepted bool   `json:"terms_accepted" binding:"required"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

// RefreshTokenRequest represents the request body for token refresh.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest represents the request body for user logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest represents the request body for forgot password.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the request body for password reset.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// AuthResponse represents the response for authentication endpoints.
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// TokenResponse represents the response for token refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// MessageResponse represents a generic message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// UserResponse represents the user data in API responses.
type UserResponse struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	Name               string    `json:"name"`
	DateFormat         string    `json:"date_format"`
	NumberFormat       string    `json:"number_format"`
	FirstDayOfWeek     string    `json:"first_day_of_week"`
	EmailNotifications bool      `json:"email_notifications"`
	GoalAlerts         bool      `json:"goal_alerts"`
	RecurringReminders bool      `json:"recurring_reminders"`
	CreatedAt          time.Time `json:"created_at"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ToUserResponse converts a domain User entity to a UserResponse DTO.
func ToUserResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:                 user.ID.String(),
		Email:              user.Email,
		Name:               user.Name,
		DateFormat:         string(user.DateFormat),
		NumberFormat:       string(user.NumberFormat),
		FirstDayOfWeek:     string(user.FirstDayOfWeek),
		EmailNotifications: user.EmailNotifications,
		GoalAlerts:         user.GoalAlerts,
		RecurringReminders: user.RecurringReminders,
		CreatedAt:          user.CreatedAt,
	}
}

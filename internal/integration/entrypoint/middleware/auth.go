// Package middleware provides HTTP middleware for the API endpoints.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
)

// ContextKey is a type for context keys.
type ContextKey string

const (
	// UserIDKey is the context key for the authenticated user's ID.
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for the authenticated user's email.
	UserEmailKey ContextKey = "user_email"
)

// AuthMiddleware provides JWT authentication middleware.
type AuthMiddleware struct {
	tokenService adapter.TokenService
}

// NewAuthMiddleware creates a new auth middleware instance.
func NewAuthMiddleware(tokenService adapter.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// Authenticate returns a Gin middleware handler that enforces JWT authentication.
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Authorization header is required",
				Code:  string(domainerror.ErrCodeMissingToken),
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Invalid authorization header format",
				Code:  string(domainerror.ErrCodeInvalidToken),
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Token is required",
				Code:  string(domainerror.ErrCodeMissingToken),
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := m.tokenService.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Invalid or expired token",
				Code:  string(domainerror.ErrCodeInvalidToken),
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UserEmailKey), claims.Email)

		c.Next()
	}
}

// GetUserIDFromContext extracts the user ID from the Gin context.
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return uuid.Nil, false
	}
	id, ok := userID.(uuid.UUID)
	return id, ok
}

// GetUserEmailFromContext extracts the user email from the Gin context.
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get(string(UserEmailKey))
	if !exists {
		return "", false
	}
	emailStr, ok := email.(string)
	return emailStr, ok
}

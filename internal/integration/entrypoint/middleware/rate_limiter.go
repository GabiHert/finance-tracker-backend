// Package middleware provides HTTP middleware for the API endpoints.
package middleware

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/dto"
)

const (
	// defaultMaxAttempts is the default number of allowed attempts per window.
	defaultMaxAttempts = 5
	// defaultWindowDuration is the default time window for rate limiting.
	defaultWindowDuration = 1 * time.Minute
)

// rateLimitEntry tracks rate limit data for a single key.
type rateLimitEntry struct {
	attempts  int
	resetTime time.Time
}

// RateLimiter provides IP-based rate limiting functionality.
type RateLimiter struct {
	mu             sync.Mutex
	entries        map[string]*rateLimitEntry
	maxAttempts    int
	windowDuration time.Duration
}

// NewRateLimiter creates a new rate limiter with default settings.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		entries:        make(map[string]*rateLimitEntry),
		maxAttempts:    defaultMaxAttempts,
		windowDuration: defaultWindowDuration,
	}
}

// NewRateLimiterWithConfig creates a new rate limiter with custom settings.
func NewRateLimiterWithConfig(maxAttempts int, windowDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		entries:        make(map[string]*rateLimitEntry),
		maxAttempts:    maxAttempts,
		windowDuration: windowDuration,
	}
}

// Middleware returns a Gin middleware handler that enforces rate limiting.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting in E2E mode or test environment
		if os.Getenv("E2E_MODE") == "true" || os.Getenv("ENV") == "test" {
			c.Next()
			return
		}

		// Get client IP
		clientIP := c.ClientIP()
		if clientIP == "" {
			clientIP = c.Request.RemoteAddr
		}

		// Check rate limit
		if !rl.allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error: "Too many requests. Please try again later.",
				Code:  string(domainerror.ErrCodeRateLimited),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if a request from the given key should be allowed.
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	entry, exists := rl.entries[key]
	if !exists {
		// First request from this key
		rl.entries[key] = &rateLimitEntry{
			attempts:  1,
			resetTime: now.Add(rl.windowDuration),
		}
		return true
	}

	// Check if the window has expired
	if now.After(entry.resetTime) {
		// Reset the window
		entry.attempts = 1
		entry.resetTime = now.Add(rl.windowDuration)
		return true
	}

	// Check if under the limit
	if entry.attempts < rl.maxAttempts {
		entry.attempts++
		return true
	}

	// Rate limit exceeded
	return false
}

// Reset clears the rate limiter state (useful for testing).
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.entries = make(map[string]*rateLimitEntry)
}

// Cleanup removes expired entries (can be called periodically to free memory).
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, entry := range rl.entries {
		if now.After(entry.resetTime) {
			delete(rl.entries, key)
		}
	}
}

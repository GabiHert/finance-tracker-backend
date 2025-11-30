// Package router sets up the HTTP routing for the application.
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/finance-tracker/backend/internal/integration/entrypoint/controller"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// Router holds the Gin engine and controller dependencies.
type Router struct {
	engine                *gin.Engine
	healthController      *controller.HealthController
	authController        *controller.AuthController
	userController        *controller.UserController
	categoryController    *controller.CategoryController
	transactionController *controller.TransactionController
	loginRateLimiter      *middleware.RateLimiter
	authMiddleware        *middleware.AuthMiddleware
}

// NewRouter creates a new router instance with all dependencies.
func NewRouter(
	healthController *controller.HealthController,
	authController *controller.AuthController,
	userController *controller.UserController,
	categoryController *controller.CategoryController,
	transactionController *controller.TransactionController,
	loginRateLimiter *middleware.RateLimiter,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		healthController:      healthController,
		authController:        authController,
		userController:        userController,
		categoryController:    categoryController,
		transactionController: transactionController,
		loginRateLimiter:      loginRateLimiter,
		authMiddleware:        authMiddleware,
	}
}

// Setup configures and returns the Gin engine with all routes.
func (r *Router) Setup(environment string) *gin.Engine {
	// Set Gin mode based on environment
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else if environment == "test" {
		gin.SetMode(gin.TestMode)
	}

	// Create router with default middleware (logger and recovery)
	r.engine = gin.Default()

	// Setup routes
	r.setupHealthRoutes()
	r.setupAPIRoutes()

	return r.engine
}

// setupHealthRoutes configures health check endpoints.
func (r *Router) setupHealthRoutes() {
	r.engine.GET("/health", r.healthController.Check)
}

// setupAPIRoutes configures the main API routes.
func (r *Router) setupAPIRoutes() {
	// API v1 group
	v1 := r.engine.Group("/api/v1")
	{
		// Auth routes (only setup if auth controller is available)
		if r.authController != nil && r.loginRateLimiter != nil {
			auth := v1.Group("/auth")
			{
				auth.POST("/register", r.authController.Register)
				auth.POST("/login", r.loginRateLimiter.Middleware(), r.authController.Login)
				auth.POST("/refresh", r.authController.RefreshToken)
				auth.POST("/logout", r.authController.Logout)
				auth.POST("/forgot-password", r.authController.ForgotPassword)
				auth.POST("/reset-password", r.authController.ResetPassword)
			}
		}

		// Category routes (require authentication)
		if r.categoryController != nil && r.authMiddleware != nil {
			categories := v1.Group("/categories")
			categories.Use(r.authMiddleware.Authenticate())
			{
				categories.GET("", r.categoryController.List)
				categories.POST("", r.categoryController.Create)
				categories.PATCH("/:id", r.categoryController.Update)
				categories.DELETE("/:id", r.categoryController.Delete)
			}
		}

		// Transaction routes (require authentication)
		if r.transactionController != nil && r.authMiddleware != nil {
			transactions := v1.Group("/transactions")
			transactions.Use(r.authMiddleware.Authenticate())
			{
				transactions.GET("", r.transactionController.List)
				transactions.POST("", r.transactionController.Create)
				transactions.PATCH("/:id", r.transactionController.Update)
				transactions.DELETE("/:id", r.transactionController.Delete)
				transactions.POST("/bulk-delete", r.transactionController.BulkDelete)
				transactions.POST("/bulk-categorize", r.transactionController.BulkCategorize)
			}
		}

		// User routes (require authentication)
		if r.userController != nil && r.authMiddleware != nil {
			users := v1.Group("/users")
			users.Use(r.authMiddleware.Authenticate())
			{
				users.DELETE("/me", r.userController.DeleteAccount)
			}
		}

		// Goal routes (to be implemented)
		_ = v1.Group("/goals")

		// Group routes (to be implemented)
		_ = v1.Group("/groups")

		// Dashboard routes (to be implemented)
		_ = v1.Group("/dashboard")
	}
}

// Engine returns the underlying Gin engine.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

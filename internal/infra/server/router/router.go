// Package router sets up the HTTP routing for the application.
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/finance-tracker/backend/internal/integration/entrypoint/controller"
)

// Router holds the Gin engine and controller dependencies.
type Router struct {
	engine           *gin.Engine
	healthController *controller.HealthController
}

// NewRouter creates a new router instance with all dependencies.
func NewRouter(healthController *controller.HealthController) *Router {
	return &Router{
		healthController: healthController,
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
		// Auth routes (to be implemented)
		_ = v1.Group("/auth")

		// User routes (to be implemented)
		_ = v1.Group("/users")

		// Transaction routes (to be implemented)
		_ = v1.Group("/transactions")

		// Category routes (to be implemented)
		_ = v1.Group("/categories")

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

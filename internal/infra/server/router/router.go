// Package router sets up the HTTP routing for the application.
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/finance-tracker/backend/internal/integration/entrypoint/controller"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
)

// Router holds the Gin engine and controller dependencies.
type Router struct {
	engine                     *gin.Engine
	healthController           *controller.HealthController
	authController             *controller.AuthController
	userController             *controller.UserController
	categoryController         *controller.CategoryController
	transactionController      *controller.TransactionController
	creditCardController       *controller.CreditCardController
	reconciliationController   *controller.ReconciliationController
	goalController             *controller.GoalController
	groupController            *controller.GroupController
	categoryRuleController     *controller.CategoryRuleController
	dashboardController        *controller.DashboardController
	aiCategorizationController *controller.AiCategorizationController
	loginRateLimiter           *middleware.RateLimiter
	authMiddleware             *middleware.AuthMiddleware
}

// NewRouter creates a new router instance with all dependencies.
func NewRouter(
	healthController *controller.HealthController,
	authController *controller.AuthController,
	userController *controller.UserController,
	categoryController *controller.CategoryController,
	transactionController *controller.TransactionController,
	creditCardController *controller.CreditCardController,
	reconciliationController *controller.ReconciliationController,
	goalController *controller.GoalController,
	groupController *controller.GroupController,
	categoryRuleController *controller.CategoryRuleController,
	dashboardController *controller.DashboardController,
	aiCategorizationController *controller.AiCategorizationController,
	loginRateLimiter *middleware.RateLimiter,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		healthController:           healthController,
		authController:             authController,
		userController:             userController,
		categoryController:         categoryController,
		transactionController:      transactionController,
		creditCardController:       creditCardController,
		reconciliationController:   reconciliationController,
		goalController:             goalController,
		groupController:            groupController,
		categoryRuleController:     categoryRuleController,
		dashboardController:        dashboardController,
		aiCategorizationController: aiCategorizationController,
		loginRateLimiter:           loginRateLimiter,
		authMiddleware:             authMiddleware,
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

				// Credit card import routes (nested under transactions)
				if r.creditCardController != nil {
					creditCard := transactions.Group("/credit-card")
					{
						creditCard.POST("/preview", r.creditCardController.Preview)
						creditCard.POST("/import", r.creditCardController.Import)
						creditCard.POST("/collapse", r.creditCardController.Collapse)
						creditCard.GET("/status", r.creditCardController.GetStatus)

						// Reconciliation routes (nested under credit-card)
						if r.reconciliationController != nil {
							reconciliation := creditCard.Group("/reconciliation")
							{
								reconciliation.GET("/pending", r.reconciliationController.GetPending)
								reconciliation.GET("/linked", r.reconciliationController.GetLinked)
								reconciliation.GET("/summary", r.reconciliationController.GetSummary)
								reconciliation.POST("/link", r.reconciliationController.ManualLink)
								reconciliation.POST("/unlink", r.reconciliationController.Unlink)
								reconciliation.POST("/trigger", r.reconciliationController.TriggerReconciliation)
							}
						}
					}
				}
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

		// Goal routes (require authentication)
		if r.goalController != nil && r.authMiddleware != nil {
			goals := v1.Group("/goals")
			goals.Use(r.authMiddleware.Authenticate())
			{
				goals.GET("", r.goalController.List)
				goals.POST("", r.goalController.Create)
				goals.GET("/:id", r.goalController.Get)
				goals.PATCH("/:id", r.goalController.Update)
				goals.DELETE("/:id", r.goalController.Delete)
			}
		}

		// Group routes (require authentication)
		if r.groupController != nil && r.authMiddleware != nil {
			groups := v1.Group("/groups")
			groups.Use(r.authMiddleware.Authenticate())
			{
				groups.POST("", r.groupController.Create)
				groups.GET("", r.groupController.List)
				groups.GET("/:id", r.groupController.Get)
			groups.DELETE("/:id", r.groupController.Delete)
				groups.GET("/:id/dashboard", r.groupController.GetDashboard)
				groups.POST("/:id/invite/check", r.groupController.CheckInvite)
				groups.POST("/:id/invite", r.groupController.Invite)
				groups.PUT("/:id/members/:member_id/role", r.groupController.ChangeRole)
				groups.DELETE("/:id/members/:member_id", r.groupController.RemoveMember)
				groups.DELETE("/:id/members/me", r.groupController.Leave)
				// Group category routes
				groups.GET("/:id/categories", r.groupController.ListCategories)
				groups.POST("/:id/categories", r.groupController.CreateCategory)
			}

			// Invite acceptance route (separate path)
			invites := v1.Group("/groups/invites")
			invites.Use(r.authMiddleware.Authenticate())
			{
				invites.POST("/:token/accept", r.groupController.AcceptInvite)
			}
		}

		// Category rule routes (require authentication)
		if r.categoryRuleController != nil && r.authMiddleware != nil {
			categoryRules := v1.Group("/category-rules")
			categoryRules.Use(r.authMiddleware.Authenticate())
			{
				categoryRules.GET("", r.categoryRuleController.List)
				categoryRules.POST("", r.categoryRuleController.Create)
				categoryRules.POST("/test", r.categoryRuleController.TestPattern)
				categoryRules.PATCH("/reorder", r.categoryRuleController.Reorder)
				categoryRules.PATCH("/:id", r.categoryRuleController.Update)
				categoryRules.DELETE("/:id", r.categoryRuleController.Delete)
			}
		}

		// Dashboard routes (require authentication)
		if r.dashboardController != nil && r.authMiddleware != nil {
			dashboard := v1.Group("/dashboard")
			dashboard.Use(r.authMiddleware.Authenticate())
			{
				dashboard.GET("/category-trends", r.dashboardController.GetCategoryTrends)
			}
		}

		// AI Categorization routes (require authentication)
		if r.aiCategorizationController != nil && r.authMiddleware != nil {
			ai := v1.Group("/ai/categorization")
			ai.Use(r.authMiddleware.Authenticate())
			{
				ai.GET("/status", r.aiCategorizationController.GetStatus)
				ai.POST("/start", r.aiCategorizationController.Start)
				ai.GET("/suggestions", r.aiCategorizationController.GetSuggestions)
				ai.POST("/suggestions/:id/approve", r.aiCategorizationController.ApproveSuggestion)
				ai.POST("/suggestions/:id/reject", r.aiCategorizationController.RejectSuggestion)
				ai.DELETE("/suggestions", r.aiCategorizationController.ClearSuggestions)
			}
		}
	}
}

// Engine returns the underlying Gin engine.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

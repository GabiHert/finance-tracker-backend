// Package dependency provides dependency injection for the application.
package dependency

import (
	"time"

	"gorm.io/gorm"

	"github.com/finance-tracker/backend/config"
	"github.com/finance-tracker/backend/internal/application/usecase/auth"
	"github.com/finance-tracker/backend/internal/application/usecase/category"
	"github.com/finance-tracker/backend/internal/application/usecase/goal"
	"github.com/finance-tracker/backend/internal/application/usecase/transaction"
	"github.com/finance-tracker/backend/internal/infra/server/router"
	"github.com/finance-tracker/backend/internal/integration/adapters"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/controller"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
	"github.com/finance-tracker/backend/internal/integration/persistence"
)

// Injector holds all application dependencies.
type Injector struct {
	Config *config.Config
	DB     *gorm.DB
	Router *router.Router
}

// NewInjector creates a new dependency injector with all dependencies wired.
func NewInjector(cfg *config.Config, db *gorm.DB) *Injector {
	// Create repositories
	userRepo := persistence.NewUserRepository(db)
	tokenRepo := persistence.NewTokenRepository(db)
	categoryRepo := persistence.NewCategoryRepository(db)
	transactionRepo := persistence.NewTransactionRepository(db)
	goalRepo := persistence.NewGoalRepository(db)

	// Create adapters/services
	passwordService := adapters.NewPasswordService()
	tokenService := adapters.NewTokenService(cfg.JWT.Secret, tokenRepo)
	resetTokenService := adapters.NewPasswordResetTokenService(tokenRepo)

	// Create auth use cases
	registerUseCase := auth.NewRegisterUserUseCase(userRepo, passwordService, tokenService)
	loginUseCase := auth.NewLoginUserUseCase(userRepo, passwordService, tokenService)
	refreshTokenUseCase := auth.NewRefreshTokenUseCase(tokenService)
	logoutUseCase := auth.NewLogoutUserUseCase(tokenService)
	forgotPasswordUseCase := auth.NewForgotPasswordUseCase(userRepo, resetTokenService)
	resetPasswordUseCase := auth.NewResetPasswordUseCase(userRepo, passwordService, resetTokenService)
	deleteAccountUseCase := auth.NewDeleteAccountUseCase(userRepo, passwordService, tokenService)

	// Create category use cases
	listCategoriesUseCase := category.NewListCategoriesUseCase(categoryRepo)
	createCategoryUseCase := category.NewCreateCategoryUseCase(categoryRepo)
	updateCategoryUseCase := category.NewUpdateCategoryUseCase(categoryRepo)
	deleteCategoryUseCase := category.NewDeleteCategoryUseCase(categoryRepo)

	// Create transaction use cases
	listTransactionsUseCase := transaction.NewListTransactionsUseCase(transactionRepo)
	createTransactionUseCase := transaction.NewCreateTransactionUseCase(transactionRepo, categoryRepo)
	updateTransactionUseCase := transaction.NewUpdateTransactionUseCase(transactionRepo, categoryRepo)
	deleteTransactionUseCase := transaction.NewDeleteTransactionUseCase(transactionRepo)
	bulkDeleteTransactionsUseCase := transaction.NewBulkDeleteTransactionsUseCase(transactionRepo)
	bulkCategorizeTransactionsUseCase := transaction.NewBulkCategorizeTransactionsUseCase(transactionRepo, categoryRepo)

	// Create goal use cases
	listGoalsUseCase := goal.NewListGoalsUseCase(goalRepo, categoryRepo)
	createGoalUseCase := goal.NewCreateGoalUseCase(goalRepo, categoryRepo)
	getGoalUseCase := goal.NewGetGoalUseCase(goalRepo, categoryRepo)
	updateGoalUseCase := goal.NewUpdateGoalUseCase(goalRepo)
	deleteGoalUseCase := goal.NewDeleteGoalUseCase(goalRepo)

	// Create controllers
	healthController := controller.NewHealthController(func() bool {
		sqlDB, err := db.DB()
		if err != nil {
			return false
		}
		return sqlDB.Ping() == nil
	})

	authController := controller.NewAuthController(
		registerUseCase,
		loginUseCase,
		refreshTokenUseCase,
		logoutUseCase,
		forgotPasswordUseCase,
		resetPasswordUseCase,
	)

	userController := controller.NewUserController(
		deleteAccountUseCase,
	)

	categoryController := controller.NewCategoryController(
		listCategoriesUseCase,
		createCategoryUseCase,
		updateCategoryUseCase,
		deleteCategoryUseCase,
	)

	transactionController := controller.NewTransactionController(
		listTransactionsUseCase,
		createTransactionUseCase,
		updateTransactionUseCase,
		deleteTransactionUseCase,
		bulkDeleteTransactionsUseCase,
		bulkCategorizeTransactionsUseCase,
	)

	goalController := controller.NewGoalController(
		listGoalsUseCase,
		createGoalUseCase,
		getGoalUseCase,
		updateGoalUseCase,
		deleteGoalUseCase,
	)

	// Create middleware
	// Use higher rate limits for E2E/test environments to prevent flaky tests
	var loginRateLimiter *middleware.RateLimiter
	if cfg.Server.Environment == "e2e" || cfg.Server.Environment == "test" {
		loginRateLimiter = middleware.NewRateLimiterWithConfig(1000, 1*time.Minute)
	} else {
		loginRateLimiter = middleware.NewRateLimiter()
	}
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	// Create router
	r := router.NewRouter(healthController, authController, userController, categoryController, transactionController, goalController, loginRateLimiter, authMiddleware)

	return &Injector{
		Config: cfg,
		DB:     db,
		Router: r,
	}
}

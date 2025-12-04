// Package dependency provides dependency injection for the application.
package dependency

import (
	"time"

	"gorm.io/gorm"

	"github.com/finance-tracker/backend/config"
	"github.com/finance-tracker/backend/internal/application/usecase/auth"
	"github.com/finance-tracker/backend/internal/application/usecase/category"
	categoryrule "github.com/finance-tracker/backend/internal/application/usecase/category_rule"
	creditcard "github.com/finance-tracker/backend/internal/application/usecase/credit_card"
	"github.com/finance-tracker/backend/internal/application/usecase/goal"
	"github.com/finance-tracker/backend/internal/application/usecase/group"
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
	groupRepo := persistence.NewGroupRepository(db)
	categoryRuleRepo := persistence.NewCategoryRuleRepository(db)

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
	createTransactionUseCase := transaction.NewCreateTransactionUseCase(transactionRepo, categoryRepo, categoryRuleRepo)
	updateTransactionUseCase := transaction.NewUpdateTransactionUseCase(transactionRepo, categoryRepo)
	deleteTransactionUseCase := transaction.NewDeleteTransactionUseCase(transactionRepo)
	bulkDeleteTransactionsUseCase := transaction.NewBulkDeleteTransactionsUseCase(transactionRepo)
	bulkCategorizeTransactionsUseCase := transaction.NewBulkCategorizeTransactionsUseCase(transactionRepo, categoryRepo)

	// Create credit card use cases
	previewImportUseCase := creditcard.NewPreviewImportUseCase(transactionRepo)
	importTransactionsUseCase := creditcard.NewImportTransactionsUseCase(transactionRepo, categoryRepo, categoryRuleRepo)
	collapseExpansionUseCase := creditcard.NewCollapseExpansionUseCase(transactionRepo)
	getStatusUseCase := creditcard.NewGetStatusUseCase(transactionRepo)

	// Create goal use cases
	listGoalsUseCase := goal.NewListGoalsUseCase(goalRepo, categoryRepo)
	createGoalUseCase := goal.NewCreateGoalUseCase(goalRepo, categoryRepo)
	getGoalUseCase := goal.NewGetGoalUseCase(goalRepo, categoryRepo)
	updateGoalUseCase := goal.NewUpdateGoalUseCase(goalRepo)
	deleteGoalUseCase := goal.NewDeleteGoalUseCase(goalRepo)

	// Create group use cases
	createGroupUseCase := group.NewCreateGroupUseCase(groupRepo, userRepo)
	listGroupsUseCase := group.NewListGroupsUseCase(groupRepo)
	getGroupUseCase := group.NewGetGroupUseCase(groupRepo)
	deleteGroupUseCase := group.NewDeleteGroupUseCase(groupRepo)
	inviteMemberUseCase := group.NewInviteMemberUseCase(groupRepo, userRepo)
	acceptInviteUseCase := group.NewAcceptInviteUseCase(groupRepo, userRepo)
	changeMemberRoleUseCase := group.NewChangeMemberRoleUseCase(groupRepo)
	removeMemberUseCase := group.NewRemoveMemberUseCase(groupRepo)
	leaveGroupUseCase := group.NewLeaveGroupUseCase(groupRepo)
	getGroupDashboardUseCase := group.NewGetGroupDashboardUseCase(groupRepo)

	// Create category rule use cases
	listCategoryRulesUseCase := categoryrule.NewListCategoryRulesUseCase(categoryRuleRepo)
	createCategoryRuleUseCase := categoryrule.NewCreateCategoryRuleUseCase(categoryRuleRepo, categoryRepo, transactionRepo)
	updateCategoryRuleUseCase := categoryrule.NewUpdateCategoryRuleUseCase(categoryRuleRepo, categoryRepo)
	deleteCategoryRuleUseCase := categoryrule.NewDeleteCategoryRuleUseCase(categoryRuleRepo)
	reorderCategoryRulesUseCase := categoryrule.NewReorderCategoryRulesUseCase(categoryRuleRepo)
	testPatternUseCase := categoryrule.NewTestPatternUseCase(categoryRuleRepo)

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

	creditCardController := controller.NewCreditCardController(
		previewImportUseCase,
		importTransactionsUseCase,
		collapseExpansionUseCase,
		getStatusUseCase,
	)

	goalController := controller.NewGoalController(
		listGoalsUseCase,
		createGoalUseCase,
		getGoalUseCase,
		updateGoalUseCase,
		deleteGoalUseCase,
	)

	groupController := controller.NewGroupController(
		createGroupUseCase,
		listGroupsUseCase,
		getGroupUseCase,
		deleteGroupUseCase,
		inviteMemberUseCase,
		acceptInviteUseCase,
		changeMemberRoleUseCase,
		removeMemberUseCase,
		leaveGroupUseCase,
		getGroupDashboardUseCase,
		listCategoriesUseCase,
		createCategoryUseCase,
		groupRepo,
	)

	categoryRuleController := controller.NewCategoryRuleController(
		listCategoryRulesUseCase,
		createCategoryRuleUseCase,
		updateCategoryRuleUseCase,
		deleteCategoryRuleUseCase,
		reorderCategoryRulesUseCase,
		testPatternUseCase,
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
	r := router.NewRouter(healthController, authController, userController, categoryController, transactionController, creditCardController, goalController, groupController, categoryRuleController, loginRateLimiter, authMiddleware)

	return &Injector{
		Config: cfg,
		DB:     db,
		Router: r,
	}
}

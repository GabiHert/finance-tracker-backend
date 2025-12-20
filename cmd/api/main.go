// Package main is the entry point for the Finance Tracker API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/finance-tracker/backend/config"
	"github.com/finance-tracker/backend/internal/application/adapter"
	aicategorization "github.com/finance-tracker/backend/internal/application/usecase/ai_categorization"
	"github.com/finance-tracker/backend/internal/application/usecase/auth"
	"github.com/finance-tracker/backend/internal/application/usecase/category"
	categoryrule "github.com/finance-tracker/backend/internal/application/usecase/category_rule"
	creditcard "github.com/finance-tracker/backend/internal/application/usecase/credit_card"
	"github.com/finance-tracker/backend/internal/application/usecase/dashboard"
	"github.com/finance-tracker/backend/internal/application/usecase/goal"
	"github.com/finance-tracker/backend/internal/application/usecase/group"
	"github.com/finance-tracker/backend/internal/application/usecase/reconciliation"
	"github.com/finance-tracker/backend/internal/application/usecase/transaction"
	"github.com/finance-tracker/backend/internal/infra/db"
	"github.com/finance-tracker/backend/internal/infra/server/router"
	"github.com/finance-tracker/backend/internal/integration/adapters"
	"github.com/finance-tracker/backend/internal/integration/email"
	"github.com/finance-tracker/backend/internal/integration/email/templates"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/controller"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
	"github.com/finance-tracker/backend/internal/integration/persistence"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
)

func main() {
	// Load .env file if it exists (development only)
	_ = godotenv.Load()

	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	slog.Info("Starting Finance Tracker API",
		"environment", cfg.Server.Environment,
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)

	// Create a cancellable context for the email worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection
	var database *db.Database
	var dbHealthChecker func() bool

	database, err := db.NewPostgresConnection(&cfg.Database)
	if err != nil {
		slog.Warn("Database connection failed, running without database",
			"error", err,
		)
		dbHealthChecker = func() bool { return false }
	} else {
		// Run database migrations
		if err := database.AutoMigrate(
			&model.UserModel{},
			&model.RefreshTokenModel{},
			&model.PasswordResetTokenModel{},
			&model.CategoryModel{},
			&model.TransactionModel{},
			&model.GoalModel{},
			&model.GroupModel{},
			&model.GroupMemberModel{},
			&model.GroupInviteModel{},
			&model.CategoryRuleModel{},
			&model.EmailQueueModel{},
			&model.AISuggestionModel{},
		); err != nil {
			slog.Error("Failed to run database migrations", "error", err)
			os.Exit(1)
		}
		slog.Info("Database migrations completed successfully")

		dbHealthChecker = database.HealthCheck
		defer func() {
			if err := database.Close(); err != nil {
				slog.Error("Failed to close database connection", "error", err)
			}
		}()
	}

	// Create health controller with database health checker
	healthController := controller.NewHealthController(dbHealthChecker)

	// Create controllers and middleware (only if database is available)
	var authController *controller.AuthController
	var userController *controller.UserController
	var categoryController *controller.CategoryController
	var transactionController *controller.TransactionController
	var creditCardController *controller.CreditCardController
	var reconciliationController *controller.ReconciliationController
	var goalController *controller.GoalController
	var groupController *controller.GroupController
	var categoryRuleController *controller.CategoryRuleController
	var dashboardController *controller.DashboardController
	var aiCategorizationController *controller.AiCategorizationController
	var loginRateLimiter *middleware.RateLimiter
	var authMiddleware *middleware.AuthMiddleware

	if database != nil {
		// Create repositories
		userRepo := persistence.NewUserRepository(database.DB())
		tokenRepo := persistence.NewTokenRepository(database.DB())
		categoryRepo := persistence.NewCategoryRepository(database.DB())
		transactionRepo := persistence.NewTransactionRepository(database.DB())
		goalRepo := persistence.NewGoalRepository(database.DB())
		groupRepo := persistence.NewGroupRepository(database.DB())
		categoryRuleRepo := persistence.NewCategoryRuleRepository(database.DB())
		emailQueueRepo := persistence.NewEmailQueueRepository(database.DB())
		aiSuggestionRepo := persistence.NewAISuggestionRepository(database.DB())

		// Create adapters/services
		passwordService := adapters.NewPasswordService()
		tokenService := adapters.NewTokenService(cfg.JWT.Secret, tokenRepo)
		resetTokenService := adapters.NewPasswordResetTokenService(tokenRepo)
		geminiService := adapters.NewGeminiService(cfg.AI.GeminiAPIKey)
		processingTracker := aicategorization.NewInMemoryProcessingTracker()

		// Create email infrastructure
		var emailService adapter.EmailService
		var emailSender adapter.EmailSender

		if cfg.Email.ResendAPIKey != "" {
			// Use Resend client for production
			emailSender = email.NewResendClient(
				cfg.Email.ResendAPIKey,
				cfg.Email.FromName,
				cfg.Email.FromEmail,
			)
			slog.Info("Resend email sender initialized")
		} else {
			// Use mock email sender for development/testing
			emailSender = email.NewMockEmailSender()
			slog.Warn("Using mock email sender (RESEND_API_KEY not set)")
		}

		// Create email service for queueing
		emailService = email.NewService(emailQueueRepo, cfg.Email.AppBaseURL)

		// Create and start email worker if enabled
		if cfg.Email.WorkerEnabled {
			templateRenderer, err := templates.NewRenderer()
			if err != nil {
				slog.Error("Failed to initialize email template renderer", "error", err)
				os.Exit(1)
			}

			workerConfig := email.WorkerConfig{
				PollInterval: cfg.Email.PollInterval,
				BatchSize:    cfg.Email.BatchSize,
			}
			emailWorker := email.NewWorker(emailQueueRepo, emailSender, templateRenderer, workerConfig)

			// Start email worker in background
			go emailWorker.Start(ctx)
			slog.Info("Email worker started",
				"poll_interval", cfg.Email.PollInterval,
				"batch_size", cfg.Email.BatchSize,
			)
		} else {
			slog.Info("Email worker disabled")
		}

		// Create auth use cases
		registerUseCase := auth.NewRegisterUserUseCase(userRepo, passwordService, tokenService)
		loginUseCase := auth.NewLoginUserUseCase(userRepo, passwordService, tokenService)
		refreshTokenUseCase := auth.NewRefreshTokenUseCase(tokenService)
		logoutUseCase := auth.NewLogoutUserUseCase(tokenService)
		forgotPasswordUseCase := auth.NewForgotPasswordUseCase(userRepo, resetTokenService, emailService, cfg.Email.AppBaseURL)
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

		// Create reconciliation repository and use cases
		reconciliationRepo := persistence.NewReconciliationRepository(database.DB())
		getPendingUseCase := reconciliation.NewGetPendingUseCase(reconciliationRepo)
		getLinkedUseCase := reconciliation.NewGetLinkedUseCase(reconciliationRepo)
		getSummaryUseCase := reconciliation.NewGetSummaryUseCase(reconciliationRepo)
		manualLinkUseCase := reconciliation.NewManualLinkUseCase(reconciliationRepo)
		unlinkUseCase := reconciliation.NewUnlinkUseCase(reconciliationRepo)
		triggerReconciliationUseCase := reconciliation.NewTriggerReconciliationUseCase(reconciliationRepo)

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
		inviteMemberUseCase := group.NewInviteMemberUseCase(groupRepo, userRepo, emailService, cfg.Email.AppBaseURL)
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

		// Create auth controller
		authController = controller.NewAuthController(
			registerUseCase,
			loginUseCase,
			refreshTokenUseCase,
			logoutUseCase,
			forgotPasswordUseCase,
			resetPasswordUseCase,
		)

		// Create user controller
		userController = controller.NewUserController(
			deleteAccountUseCase,
		)

		// Create category controller
		categoryController = controller.NewCategoryController(
			listCategoriesUseCase,
			createCategoryUseCase,
			updateCategoryUseCase,
			deleteCategoryUseCase,
		)

		// Create transaction controller
		transactionController = controller.NewTransactionController(
			listTransactionsUseCase,
			createTransactionUseCase,
			updateTransactionUseCase,
			deleteTransactionUseCase,
			bulkDeleteTransactionsUseCase,
			bulkCategorizeTransactionsUseCase,
		)

		// Create credit card controller
		creditCardController = controller.NewCreditCardController(
			previewImportUseCase,
			importTransactionsUseCase,
			collapseExpansionUseCase,
			getStatusUseCase,
		)

		// Create reconciliation controller
		reconciliationController = controller.NewReconciliationController(
			getPendingUseCase,
			getLinkedUseCase,
			getSummaryUseCase,
			manualLinkUseCase,
			unlinkUseCase,
			triggerReconciliationUseCase,
		)

		// Create goal controller
		goalController = controller.NewGoalController(
			listGoalsUseCase,
			createGoalUseCase,
			getGoalUseCase,
			updateGoalUseCase,
			deleteGoalUseCase,
		)

		// Create group controller
		groupController = controller.NewGroupController(
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

		// Create category rule controller
		categoryRuleController = controller.NewCategoryRuleController(
			listCategoryRulesUseCase,
			createCategoryRuleUseCase,
			updateCategoryRuleUseCase,
			deleteCategoryRuleUseCase,
			reorderCategoryRulesUseCase,
			testPatternUseCase,
		)

		// Create dashboard repository and use cases
		dashboardRepo := persistence.NewDashboardRepository(database.DB())
		getCategoryTrendsUseCase := dashboard.NewGetCategoryTrendsUseCase(transactionRepo)
		getDataRangeUseCase := dashboard.NewGetDataRangeUseCase(dashboardRepo)
		getTrendsUseCase := dashboard.NewGetTrendsUseCase(dashboardRepo)
		getCategoryBreakdownUseCase := dashboard.NewGetCategoryBreakdownUseCase(dashboardRepo)
		getPeriodTransactionsUseCase := dashboard.NewGetPeriodTransactionsUseCase(dashboardRepo)

		// Create dashboard controller
		dashboardController = controller.NewDashboardController(
			getCategoryTrendsUseCase,
			getDataRangeUseCase,
			getTrendsUseCase,
			getCategoryBreakdownUseCase,
			getPeriodTransactionsUseCase,
		)

		// Create AI categorization use cases
		aiGetStatusUseCase := aicategorization.NewGetStatusUseCase(transactionRepo, aiSuggestionRepo, processingTracker)
		aiStartCategorizationUseCase := aicategorization.NewStartCategorizationUseCase(transactionRepo, categoryRepo, aiSuggestionRepo, geminiService, processingTracker)
		aiGetSuggestionsUseCase := aicategorization.NewGetSuggestionsUseCase(aiSuggestionRepo)
		aiApproveSuggestionUseCase := aicategorization.NewApproveSuggestionUseCase(aiSuggestionRepo, categoryRepo, transactionRepo, categoryRuleRepo)
		aiRejectSuggestionUseCase := aicategorization.NewRejectSuggestionUseCase(aiSuggestionRepo, geminiService, transactionRepo, categoryRepo)
		aiClearSuggestionsUseCase := aicategorization.NewClearSuggestionsUseCase(aiSuggestionRepo)

		// Create AI categorization controller
		aiCategorizationController = controller.NewAiCategorizationController(
			aiGetStatusUseCase,
			aiStartCategorizationUseCase,
			aiGetSuggestionsUseCase,
			aiApproveSuggestionUseCase,
			aiRejectSuggestionUseCase,
			aiClearSuggestionsUseCase,
		)

		// Create middleware
		loginRateLimiter = middleware.NewRateLimiter()
		authMiddleware = middleware.NewAuthMiddleware(tokenService)

		slog.Info("Auth, Category, Transaction, Goal, Group, CategoryRule, Dashboard, AI Categorization, and Email systems initialized successfully")
	} else {
		slog.Warn("Auth, Category, Transaction, Goal, Group, and Email systems not initialized due to missing database connection")
	}

	// Setup router
	r := router.NewRouter(healthController, authController, userController, categoryController, transactionController, creditCardController, reconciliationController, goalController, groupController, categoryRuleController, dashboardController, aiCategorizationController, loginRateLimiter, authMiddleware)
	engine := r.Setup(cfg.Server.Environment)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Server listening", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Cancel context to stop email worker
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited properly")
}

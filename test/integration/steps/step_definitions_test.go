package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/application/usecase/auth"
	"github.com/finance-tracker/backend/internal/application/usecase/category"
	categoryrule "github.com/finance-tracker/backend/internal/application/usecase/category_rule"
	"github.com/finance-tracker/backend/internal/application/usecase/dashboard"
	"github.com/finance-tracker/backend/internal/application/usecase/goal"
	"github.com/finance-tracker/backend/internal/application/usecase/group"
	"github.com/finance-tracker/backend/internal/application/usecase/transaction"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
	"github.com/finance-tracker/backend/internal/infra/server/router"
	"github.com/finance-tracker/backend/internal/integration/adapters"
	"github.com/finance-tracker/backend/internal/integration/email"
	"github.com/finance-tracker/backend/internal/integration/email/templates"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/controller"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/middleware"
	"github.com/finance-tracker/backend/internal/integration/persistence"
	"github.com/finance-tracker/backend/internal/integration/persistence/model"
	"github.com/finance-tracker/backend/test/integration/mock"
)

const testJWTSecret = "test-jwt-secret-key-for-testing-purposes"

var tags string

func init() {
	flag.StringVar(&tags, "scenarios", "", "tags to run")
}

func TestFeatures(t *testing.T) {
	flag.Parse()

	suite := godog.TestSuite{
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			InitializeScenario(s)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			Tags:     tags,
			Strict:   true,
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

type testContext struct {
	uri                string
	headers            map[string]string
	client             *http.Client
	response           *response
	db                 *mock.Db
	timeMock           *mock.Time
	serverPort         int
	accessToken        string
	refreshToken       string
	resetToken         string
	expiredToken       string
	currentUserID      uuid.UUID
	currentCategoryID  uuid.UUID
	currentGoalID      uuid.UUID
	currentGroupID     uuid.UUID
	currentMemberID    uuid.UUID
	currentInviteToken string
	transactionIDs     []uuid.UUID
	lastTransactionID  uuid.UUID
	// Email testing
	lastEmailJobID     uuid.UUID
	emailSenderMock    *mockEmailSender
}

type response struct {
	status int
	body   any
	err    error
}

var serverInit sync.Once
var testDB *mock.Db
var testServerPort int
var portInit sync.Once

func initializePort() {
	portInit.Do(func() {
		testServerPort = findAvailablePort()
		_ = os.Setenv("SERVER_PORT", strconv.Itoa(testServerPort))
		_ = os.Setenv("ENV", "test")
	})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	initializePort()

	test := &testContext{
		uri:        fmt.Sprintf("http://localhost:%d", testServerPort),
		client:     &http.Client{Timeout: 10 * time.Second},
		timeMock:   mock.NewTime(),
		serverPort: testServerPort,
		db: mock.NewDb("finance_tracker", map[string]any{
			"users":                 &model.UserModel{},
			"refresh_tokens":        &model.RefreshTokenModel{},
			"password_reset_tokens": &model.PasswordResetTokenModel{},
			"categories":            &model.CategoryModel{},
			"transactions":          &model.TransactionModel{},
			"goals":                 &model.GoalModel{},
			"groups":                &model.GroupModel{},
			"group_members":         &model.GroupMemberModel{},
			"group_invites":         &model.GroupInviteModel{},
			"email_queue":           &model.EmailQueueModel{},
		}),
	}

	testDB = test.db

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		test.before()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		return nil, nil
	})

	// Background steps
	ctx.Given(`^the API server is running$`, test.theAPIServerIsRunning)

	// User setup steps
	ctx.Given(`^a user exists with email "([^"]*)"$`, test.aUserExistsWithEmail)
	ctx.Given(`^a user exists with email "([^"]*)" and password "([^"]*)"$`, test.aUserExistsWithEmailAndPassword)
	ctx.Given(`^the user is logged in with valid tokens$`, test.theUserIsLoggedInWithValidTokens)
	ctx.Given(`^a password reset token exists for "([^"]*)"$`, test.aPasswordResetTokenExistsFor)
	ctx.Given(`^an expired password reset token exists$`, test.anExpiredPasswordResetTokenExists)

	// Category setup steps
	ctx.Given(`^a category exists with name "([^"]*)" and type "([^"]*)"$`, test.aCategoryExistsWithNameAndType)

	// Goal setup steps
	ctx.Given(`^a goal exists for category "([^"]*)" with limit "([^"]*)"$`, test.aGoalExistsForCategoryWithLimit)

	// Category trends setup steps
	ctx.Given(`^expense transactions exist for category trends testing$`, test.expenseTransactionsExistForCategoryTrendsTesting)

	// Group setup steps
	ctx.Given(`^I am logged in as "([^"]*)"$`, test.iAmLoggedInAs)
	ctx.Given(`^the user "([^"]*)" exists$`, test.theUserExists)

	// Header steps
	ctx.Given(`^the header is empty$`, test.theHeaderIsEmpty)
	ctx.Given(`^the header contains the key "([^"]*)" with "([^"]*)"$`, test.theHeaderContainsTheKeyWith)

	// Request steps
	ctx.When(`^I send a "([^"]*)" request to "([^"]*)"$`, test.iSendARequestTo)
	ctx.When(`^I send a "([^"]*)" request to "([^"]*)" with body:$`, test.iSendARequestToWithBody)

	// Response assertion steps
	ctx.Then(`^the response status should be (\d+)$`, test.theResponseStatusShouldBe)
	ctx.Then(`^the response should be JSON$`, test.theResponseShouldBeJSON)
	ctx.Then(`^the response should contain "([^"]*)"$`, test.theResponseShouldContain)
	ctx.Then(`^the response field "([^"]*)" should be "([^"]*)"$`, test.theResponseFieldShouldBe)
	ctx.Then(`^the response field "([^"]*)" should exist$`, test.theResponseFieldShouldExist)

	// Database assertion steps
	ctx.Then(`^the db should contain (\d+) objects in the "([^"]*)" table$`, test.theDbShouldContainObjectsInTheTable)
	ctx.Then(`^the db should contain (\d+) objects in "([^"]*)" with the values$`, test.theDbShouldContainObjectsInWithTheValues)

	// Email notification steps
	ctx.Given(`^a pending email job exists for "([^"]*)"$`, test.aPendingEmailJobExistsFor)
	ctx.Given(`^an email job with (\d+) failed attempts exists for "([^"]*)"$`, test.anEmailJobWithFailedAttemptsExistsFor)
	ctx.Given(`^the email sender will fail with a temporary error$`, test.theEmailSenderWillFailWithATemporaryError)
	ctx.Given(`^the email sender will fail with a permanent error$`, test.theEmailSenderWillFailWithAPermanentError)
	ctx.When(`^the email worker processes the queue$`, test.theEmailWorkerProcessesTheQueue)
	ctx.Then(`^an email job should be queued for "([^"]*)"$`, test.anEmailJobShouldBeQueuedFor)
	ctx.Then(`^no email job should be queued for "([^"]*)"$`, test.noEmailJobShouldBeQueuedFor)
	ctx.Then(`^the email job should have template type "([^"]*)"$`, test.theEmailJobShouldHaveTemplateType)
	ctx.Then(`^the email job should have status "([^"]*)"$`, test.theEmailJobShouldHaveStatus)
	ctx.Then(`^the email job for "([^"]*)" should have status "([^"]*)"$`, test.theEmailJobForShouldHaveStatus)
	ctx.Then(`^the email job should have a resend_id$`, test.theEmailJobShouldHaveAResendId)
	ctx.Then(`^the email job should have attempts equal to (\d+)$`, test.theEmailJobShouldHaveAttemptsEqualTo)
	ctx.Then(`^the email job should have scheduled_at in the future$`, test.theEmailJobShouldHaveScheduledAtInTheFuture)
	ctx.Then(`^the email job should have a last_error$`, test.theEmailJobShouldHaveALastError)
	ctx.Then(`^the email job template data should contain "([^"]*)"$`, test.theEmailJobTemplateDataShouldContain)
	ctx.Then(`^the database should contain an email_queue record with:$`, test.theDatabaseShouldContainAnEmailQueueRecordWith)
}

func findAvailablePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func (t *testContext) before() {
	t.headers = make(map[string]string)
	t.accessToken = ""
	t.refreshToken = ""
	t.resetToken = ""
	t.expiredToken = ""
	t.currentUserID = uuid.Nil
	t.currentCategoryID = uuid.Nil
	t.currentGoalID = uuid.Nil
	t.currentGroupID = uuid.Nil
	t.currentMemberID = uuid.Nil
	t.currentInviteToken = ""
	t.transactionIDs = nil
	t.lastTransactionID = uuid.Nil

	if t.db != nil {
		_ = t.db.ClearDB()
	}
}

func (t *testContext) startServer() {
	serverInit.Do(func() {
		go func() {
			gin.SetMode(gin.TestMode)

			// Create repositories
			userRepo := persistence.NewUserRepository(testDB.DbConn)
			tokenRepo := persistence.NewTokenRepository(testDB.DbConn)
			categoryRepo := persistence.NewCategoryRepository(testDB.DbConn)
			transactionRepo := persistence.NewTransactionRepository(testDB.DbConn)
			goalRepo := persistence.NewGoalRepository(testDB.DbConn)
			groupRepo := persistence.NewGroupRepository(testDB.DbConn)
			categoryRuleRepo := persistence.NewCategoryRuleRepository(testDB.DbConn)

			// Create adapters/services
			passwordService := adapters.NewPasswordService()
			tokenService := adapters.NewTokenService("test-jwt-secret-key-for-testing-purposes", tokenRepo)
			resetTokenService := adapters.NewPasswordResetTokenService(tokenRepo)

			// Create email queue repository and email service
			emailQueueRepo := persistence.NewEmailQueueRepository(testDB.DbConn)
			emailService := email.NewService(emailQueueRepo, "http://localhost:3000")

			// Create auth use cases (with email service for integration tests)
			registerUseCase := auth.NewRegisterUserUseCase(userRepo, passwordService, tokenService)
			loginUseCase := auth.NewLoginUserUseCase(userRepo, passwordService, tokenService)
			refreshTokenUseCase := auth.NewRefreshTokenUseCase(tokenService)
			logoutUseCase := auth.NewLogoutUserUseCase(tokenService)
			forgotPasswordUseCase := auth.NewForgotPasswordUseCase(userRepo, resetTokenService, emailService, "http://localhost:3000")
			resetPasswordUseCase := auth.NewResetPasswordUseCase(userRepo, passwordService, resetTokenService)

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

			// Create goal use cases
			listGoalsUseCase := goal.NewListGoalsUseCase(goalRepo, categoryRepo)
			createGoalUseCase := goal.NewCreateGoalUseCase(goalRepo, categoryRepo)
			getGoalUseCase := goal.NewGetGoalUseCase(goalRepo, categoryRepo)
			updateGoalUseCase := goal.NewUpdateGoalUseCase(goalRepo)
			deleteGoalUseCase := goal.NewDeleteGoalUseCase(goalRepo)

			// Create group use cases (with email service for integration tests)
			createGroupUseCase := group.NewCreateGroupUseCase(groupRepo, userRepo)
			listGroupsUseCase := group.NewListGroupsUseCase(groupRepo)
			getGroupUseCase := group.NewGetGroupUseCase(groupRepo)
			deleteGroupUseCase := group.NewDeleteGroupUseCase(groupRepo)
			inviteMemberUseCase := group.NewInviteMemberUseCase(groupRepo, userRepo, emailService, "http://localhost:3000")
			acceptInviteUseCase := group.NewAcceptInviteUseCase(groupRepo, userRepo)
			changeMemberRoleUseCase := group.NewChangeMemberRoleUseCase(groupRepo)
			removeMemberUseCase := group.NewRemoveMemberUseCase(groupRepo)
			leaveGroupUseCase := group.NewLeaveGroupUseCase(groupRepo)
			getGroupDashboardUseCase := group.NewGetGroupDashboardUseCase(groupRepo)

			// Create category rule use cases (categoryRuleRepo already created above)
			listCategoryRulesUseCase := categoryrule.NewListCategoryRulesUseCase(categoryRuleRepo)
			createCategoryRuleUseCase := categoryrule.NewCreateCategoryRuleUseCase(categoryRuleRepo, categoryRepo, transactionRepo)
			updateCategoryRuleUseCase := categoryrule.NewUpdateCategoryRuleUseCase(categoryRuleRepo, categoryRepo)
			deleteCategoryRuleUseCase := categoryrule.NewDeleteCategoryRuleUseCase(categoryRuleRepo)
			reorderCategoryRulesUseCase := categoryrule.NewReorderCategoryRulesUseCase(categoryRuleRepo)
			testPatternUseCase := categoryrule.NewTestPatternUseCase(categoryRuleRepo)

			// Create user use cases (delete account)
			deleteAccountUseCase := auth.NewDeleteAccountUseCase(userRepo, passwordService, tokenService)

			// Create controllers
			healthController := controller.NewHealthController(func() bool {
				return testDB != nil && testDB.DbConn != nil
			})

			authController := controller.NewAuthController(
				registerUseCase,
				loginUseCase,
				refreshTokenUseCase,
				logoutUseCase,
				forgotPasswordUseCase,
				resetPasswordUseCase,
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

			userController := controller.NewUserController(deleteAccountUseCase)

			// Create dashboard repository and use cases
			dashboardRepo := persistence.NewDashboardRepository(testDB.DbConn)
			getCategoryTrendsUseCase := dashboard.NewGetCategoryTrendsUseCase(transactionRepo)
			getDataRangeUseCase := dashboard.NewGetDataRangeUseCase(dashboardRepo)
			getTrendsUseCase := dashboard.NewGetTrendsUseCase(dashboardRepo)
			getCategoryBreakdownUseCase := dashboard.NewGetCategoryBreakdownUseCase(dashboardRepo)
			getPeriodTransactionsUseCase := dashboard.NewGetPeriodTransactionsUseCase(dashboardRepo)

			// Create dashboard controller
			dashboardController := controller.NewDashboardController(
				getCategoryTrendsUseCase,
				getDataRangeUseCase,
				getTrendsUseCase,
				getCategoryBreakdownUseCase,
				getPeriodTransactionsUseCase,
			)

			// Create middleware
			loginRateLimiter := middleware.NewRateLimiter()
			authMiddleware := middleware.NewAuthMiddleware(tokenService)

			r := router.NewRouter(healthController, authController, userController, categoryController, transactionController, nil, nil, goalController, groupController, categoryRuleController, dashboardController, nil, loginRateLimiter, authMiddleware)
			engine := r.Setup("test")

			addr := fmt.Sprintf(":%d", testServerPort)
			server := &http.Server{
				Addr:    addr,
				Handler: engine,
			}

			_ = server.ListenAndServe()
		}()
	})

	// Wait for server to be ready
	for i := 0; i < 50; i++ {
		resp, err := http.Get(t.uri + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (t *testContext) theAPIServerIsRunning() error {
	t.startServer()
	return nil
}

func (t *testContext) aUserExistsWithEmail(email string) error {
	return t.createUser(email, "DefaultPass123!", "Test User")
}

func (t *testContext) aUserExistsWithEmailAndPassword(email, password string) error {
	return t.createUser(email, password, "Test User")
}

func (t *testContext) createUser(email, password, name string) error {
	userID := uuid.New()
	t.currentUserID = userID

	user := &model.UserModel{
		ID:                 userID,
		Email:              email,
		Name:               name,
		PasswordHash:       hashPassword(password),
		DateFormat:         "YYYY-MM-DD",
		NumberFormat:       "US",
		FirstDayOfWeek:     "sunday",
		EmailNotifications: true,
		GoalAlerts:         true,
		RecurringReminders: true,
		TermsAcceptedAt:    time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	result := t.db.DbConn.Create(user)
	return result.Error
}

func hashPassword(password string) string {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("failed to hash password: %v", err))
	}
	return string(hashedBytes)
}

func (t *testContext) theUserIsLoggedInWithValidTokens() error {
	// Generate valid JWT tokens
	now := time.Now().UTC()

	// Generate access token
	accessClaims := jwt.MapClaims{
		"user_id":    t.currentUserID.String(),
		"email":      "test@example.com",
		"token_type": "access",
		"exp":        jwt.NewNumericDate(now.Add(15 * time.Minute)),
		"iat":        jwt.NewNumericDate(now),
		"nbf":        jwt.NewNumericDate(now),
		"iss":        "finance-tracker",
		"sub":        t.currentUserID.String(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(testJWTSecret))
	if err != nil {
		return fmt.Errorf("failed to generate access token: %w", err)
	}
	t.accessToken = accessTokenString

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"user_id":    t.currentUserID.String(),
		"email":      "test@example.com",
		"token_type": "refresh",
		"exp":        jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		"iat":        jwt.NewNumericDate(now),
		"nbf":        jwt.NewNumericDate(now),
		"iss":        "finance-tracker",
		"sub":        t.currentUserID.String(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(testJWTSecret))
	if err != nil {
		return fmt.Errorf("failed to generate refresh token: %w", err)
	}
	t.refreshToken = refreshTokenString

	// Store refresh token in database
	refreshTokenModel := &model.RefreshTokenModel{
		ID:          uuid.New(),
		Token:       t.refreshToken,
		UserID:      t.currentUserID,
		Invalidated: false,
		ExpiresAt:   now.Add(7 * 24 * time.Hour),
		CreatedAt:   now,
	}

	result := t.db.DbConn.Create(refreshTokenModel)
	return result.Error
}

func (t *testContext) aPasswordResetTokenExistsFor(email string) error {
	t.resetToken = fmt.Sprintf("test-reset-token-%s", uuid.New().String())

	var user model.UserModel
	if err := t.db.DbConn.Where("email = ?", email).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	resetTokenModel := &model.PasswordResetTokenModel{
		ID:        uuid.New(),
		Token:     t.resetToken,
		UserID:    user.ID,
		Email:     email,
		Used:      false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	result := t.db.DbConn.Create(resetTokenModel)
	return result.Error
}

func (t *testContext) anExpiredPasswordResetTokenExists() error {
	t.expiredToken = fmt.Sprintf("expired-reset-token-%s", uuid.New().String())

	resetTokenModel := &model.PasswordResetTokenModel{
		ID:        uuid.New(),
		Token:     t.expiredToken,
		UserID:    uuid.New(),
		Email:     "expired@example.com",
		Used:      false,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	result := t.db.DbConn.Create(resetTokenModel)
	return result.Error
}

func (t *testContext) theHeaderIsEmpty() error {
	t.headers = make(map[string]string)
	t.accessToken = "" // Clear access token to simulate unauthenticated request
	return nil
}

func (t *testContext) theHeaderContainsTheKeyWith(key, value string) error {
	t.headers[key] = value
	return nil
}

func (t *testContext) iSendARequestTo(method, path string) error {
	// Replace placeholders in path
	path = t.replaceTokenPlaceholders(path)
	return t.executeRequest(method, path, nil)
}

func (t *testContext) iSendARequestToWithBody(method, path string, body *godog.DocString) error {
	// Replace placeholders in path
	path = t.replaceTokenPlaceholders(path)

	var payload []byte
	if body != nil && body.Content != "" {
		content := t.replaceTokenPlaceholders(body.Content)
		payload = []byte(content)
	}
	return t.executeRequest(method, path, payload)
}

func (t *testContext) replaceTokenPlaceholders(content string) string {
	content = strings.ReplaceAll(content, "{{refresh_token}}", t.refreshToken)
	content = strings.ReplaceAll(content, "{{access_token}}", t.accessToken)
	content = strings.ReplaceAll(content, "{{reset_token}}", t.resetToken)
	content = strings.ReplaceAll(content, "{{expired_reset_token}}", t.expiredToken)
	content = strings.ReplaceAll(content, "{{category_id}}", t.currentCategoryID.String())
	content = strings.ReplaceAll(content, "{{goal_id}}", t.currentGoalID.String())
	content = strings.ReplaceAll(content, "{{transaction_id}}", t.lastTransactionID.String())
	content = strings.ReplaceAll(content, "{{group_id}}", t.currentGroupID.String())
	content = strings.ReplaceAll(content, "{{member_id}}", t.currentMemberID.String())
	content = strings.ReplaceAll(content, "{{invite_token}}", t.currentInviteToken)

	// Handle transaction_ids array placeholder
	if len(t.transactionIDs) > 0 {
		ids := make([]string, len(t.transactionIDs))
		for i, id := range t.transactionIDs {
			ids[i] = fmt.Sprintf(`"%s"`, id.String())
		}
		content = strings.ReplaceAll(content, "{{transaction_ids}}", "["+strings.Join(ids, ", ")+"]")
	}

	return content
}

func (t *testContext) executeRequest(method, path string, payload []byte) error {
	var req *http.Request
	var err error

	url := t.uri + path

	if payload != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(payload))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if t.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+t.accessToken)
	}

	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	t.response = &response{
		status: resp.StatusCode,
	}

	var responseBody map[string]any
	if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
		t.response.body = string(bodyBytes)
	} else {
		t.response.body = responseBody

		// Capture transaction ID from response if present
		if idStr, ok := responseBody["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				t.lastTransactionID = id
				t.transactionIDs = append(t.transactionIDs, id)
				// Also capture as group_id ONLY for group responses (has "name" and "members" or "created_by")
				if _, hasName := responseBody["name"]; hasName {
					if _, hasMembers := responseBody["members"]; hasMembers {
						t.currentGroupID = id
					} else if _, hasCreatedBy := responseBody["created_by"]; hasCreatedBy {
						t.currentGroupID = id
					}
				}
			}
		}

		// Capture invite token from response if present
		if token, ok := responseBody["token"].(string); ok && token != "" {
			t.currentInviteToken = token
		}

		// For accept invite response, get group details to capture member_id
		if groupID, ok := responseBody["group_id"].(string); ok {
			if gid, err := uuid.Parse(groupID); err == nil {
				t.currentGroupID = gid
				// After accepting invite, query for member details
				// This will be the newly joined member
				t.fetchMemberIDAfterAcceptInvite()
			}
		}

		// Capture member ID from group details response
		if members, ok := responseBody["members"].([]any); ok && len(members) > 0 {
			// Get the last added member (non-admin in most cases)
			for i := len(members) - 1; i >= 0; i-- {
				if member, ok := members[i].(map[string]any); ok {
					if role, ok := member["role"].(string); ok && role == "member" {
						if idStr, ok := member["id"].(string); ok {
							if id, err := uuid.Parse(idStr); err == nil {
								t.currentMemberID = id
								break
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (t *testContext) theResponseStatusShouldBe(expectedStatus int) error {
	if t.response == nil {
		return errors.New("no response received")
	}
	if t.response.status != expectedStatus {
		return fmt.Errorf("expected status %d, got %d (body: %v)", expectedStatus, t.response.status, t.response.body)
	}
	return nil
}

func (t *testContext) theResponseShouldBeJSON() error {
	if t.response == nil {
		return errors.New("no response received")
	}
	if _, ok := t.response.body.(map[string]any); !ok {
		return fmt.Errorf("response is not JSON: %v", t.response.body)
	}
	return nil
}

func (t *testContext) theResponseShouldContain(field string) error {
	if t.response == nil {
		return errors.New("no response received")
	}

	body, ok := t.response.body.(map[string]any)
	if !ok {
		return fmt.Errorf("response is not a JSON object: %v", t.response.body)
	}

	if _, exists := body[field]; !exists {
		return fmt.Errorf("response does not contain field '%s': %v", field, body)
	}
	return nil
}

func (t *testContext) theResponseFieldShouldBe(field, expectedValue string) error {
	if t.response == nil {
		return errors.New("no response received")
	}

	body, ok := t.response.body.(map[string]any)
	if !ok {
		return fmt.Errorf("response is not a JSON object: %v", t.response.body)
	}

	value := getFieldValue(body, field)
	if value == nil {
		return fmt.Errorf("field '%s' not found in response: %v", field, body)
	}

	actualValue := fmt.Sprintf("%v", value)
	if actualValue != expectedValue {
		return fmt.Errorf("field '%s' expected '%s', got '%s'", field, expectedValue, actualValue)
	}
	return nil
}

func (t *testContext) theResponseFieldShouldExist(field string) error {
	if t.response == nil {
		return errors.New("no response received")
	}

	body, ok := t.response.body.(map[string]any)
	if !ok {
		return fmt.Errorf("response is not a JSON object: %v", t.response.body)
	}

	value := getFieldValue(body, field)
	if value == nil {
		return fmt.Errorf("field '%s' not found in response: %v", field, body)
	}
	return nil
}

func (t *testContext) theDbShouldContainObjectsInTheTable(quantity int, table string) error {
	if entity, ok := t.db.GetModel(table); ok {
		entityType := reflect.TypeOf(entity).Elem()
		entitySlice := reflect.MakeSlice(reflect.SliceOf(entityType), 0, 0)
		entitySlicePtr := reflect.New(entitySlice.Type())
		entitySlicePtr.Elem().Set(entitySlice)

		result := t.db.DbConn.Unscoped().Find(entitySlicePtr.Interface())
		if result.Error != nil {
			return result.Error
		}

		count := entitySlicePtr.Elem().Len()
		if count != quantity {
			return fmt.Errorf("expected %d objects in '%s', got %d", quantity, table, count)
		}
		return nil
	}
	return fmt.Errorf("table '%s' not found in models", table)
}

func (t *testContext) theDbShouldContainObjectsInWithTheValues(quantity int, table string, content *godog.DocString) error {
	var criteria map[string]any
	if err := json.Unmarshal([]byte(content.Content), &criteria); err != nil {
		return err
	}

	if entity, ok := t.db.GetModel(table); ok {
		entityType := reflect.TypeOf(entity).Elem()
		entitySlice := reflect.MakeSlice(reflect.SliceOf(entityType), 0, 0)
		entitySlicePtr := reflect.New(entitySlice.Type())
		entitySlicePtr.Elem().Set(entitySlice)

		query := t.db.DbConn.Unscoped()
		for key, value := range criteria {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}

		result := query.Find(entitySlicePtr.Interface())
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		count := entitySlicePtr.Elem().Len()
		if count != quantity {
			return fmt.Errorf("expected %d objects in '%s' with criteria %v, got %d", quantity, table, criteria, count)
		}
		return nil
	}
	return fmt.Errorf("table '%s' not found in models", table)
}

func getFieldValue(object any, dotSeparatedField string) any {
	if object == nil {
		return nil
	}

	var objectMap map[string]any
	switch v := object.(type) {
	case map[string]any:
		objectMap = v
	default:
		objectJSON, _ := json.Marshal(object)
		if err := json.Unmarshal(objectJSON, &objectMap); err != nil {
			return nil
		}
	}

	fields := strings.Split(dotSeparatedField, ".")
	var field any = objectMap

	for _, currentField := range fields {
		if field == nil {
			return nil
		}

		if i, err := strconv.Atoi(currentField); err == nil {
			if arr, ok := field.([]any); ok && i < len(arr) {
				field = arr[i]
			} else {
				return nil
			}
		} else {
			if m, ok := field.(map[string]any); ok {
				field = m[currentField]
			} else {
				return nil
			}
		}
	}

	return field
}

// aCategoryExistsWithNameAndType creates a category with the given name and type.
func (t *testContext) aCategoryExistsWithNameAndType(name, categoryType string) error {
	categoryID := uuid.New()
	t.currentCategoryID = categoryID

	now := time.Now().UTC()
	categoryModel := &model.CategoryModel{
		ID:        categoryID,
		Name:      name,
		Color:     "#6366F1",
		Icon:      "tag",
		OwnerType: "user",
		OwnerID:   t.currentUserID,
		Type:      categoryType,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := t.db.DbConn.Create(categoryModel)
	return result.Error
}

// theUserExists creates a user with the given email if they don't already exist.
func (t *testContext) theUserExists(email string) error {
	var userModel model.UserModel
	if err := t.db.DbConn.Where("email = ?", email).First(&userModel).Error; err == nil {
		// User already exists
		return nil
	}

	// Create the user
	userID := uuid.New()
	user := &model.UserModel{
		ID:                 userID,
		Email:              email,
		Name:               "Test User " + email,
		PasswordHash:       hashPassword("SecurePass123!"),
		DateFormat:         "YYYY-MM-DD",
		NumberFormat:       "US",
		FirstDayOfWeek:     "sunday",
		EmailNotifications: true,
		GoalAlerts:         true,
		RecurringReminders: true,
		TermsAcceptedAt:    time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	result := t.db.DbConn.Create(user)
	return result.Error
}

// iAmLoggedInAs switches the current logged in user to the specified email.
func (t *testContext) iAmLoggedInAs(email string) error {
	// First ensure the user exists
	if err := t.theUserExists(email); err != nil {
		return err
	}

	// Find the user
	var userModel model.UserModel
	if err := t.db.DbConn.Where("email = ?", email).First(&userModel).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update current user ID
	t.currentUserID = userModel.ID

	// Generate new tokens for this user
	now := time.Now().UTC()

	// Generate access token
	accessClaims := jwt.MapClaims{
		"user_id":    t.currentUserID.String(),
		"email":      email,
		"token_type": "access",
		"exp":        jwt.NewNumericDate(now.Add(15 * time.Minute)),
		"iat":        jwt.NewNumericDate(now),
		"nbf":        jwt.NewNumericDate(now),
		"iss":        "finance-tracker",
		"sub":        t.currentUserID.String(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(testJWTSecret))
	if err != nil {
		return fmt.Errorf("failed to generate access token: %w", err)
	}
	t.accessToken = accessTokenString

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"user_id":    t.currentUserID.String(),
		"email":      email,
		"token_type": "refresh",
		"exp":        jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		"iat":        jwt.NewNumericDate(now),
		"nbf":        jwt.NewNumericDate(now),
		"iss":        "finance-tracker",
		"sub":        t.currentUserID.String(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(testJWTSecret))
	if err != nil {
		return fmt.Errorf("failed to generate refresh token: %w", err)
	}
	t.refreshToken = refreshTokenString

	// Check if refresh token already exists and update it, otherwise create
	var existingToken model.RefreshTokenModel
	if err := t.db.DbConn.Where("user_id = ?", t.currentUserID).First(&existingToken).Error; err == nil {
		// Update existing token
		existingToken.Token = t.refreshToken
		existingToken.Invalidated = false
		existingToken.ExpiresAt = now.Add(7 * 24 * time.Hour)
		return t.db.DbConn.Save(&existingToken).Error
	}

	// Create new token
	refreshTokenModel := &model.RefreshTokenModel{
		ID:          uuid.New(),
		Token:       t.refreshToken,
		UserID:      t.currentUserID,
		Invalidated: false,
		ExpiresAt:   now.Add(7 * 24 * time.Hour),
		CreatedAt:   now,
	}

	result := t.db.DbConn.Create(refreshTokenModel)
	return result.Error
}

// fetchMemberIDAfterAcceptInvite queries the database to get the member ID of the newly joined member.
func (t *testContext) fetchMemberIDAfterAcceptInvite() {
	// Query the database directly for the member with role "member" in the current group
	var memberModel model.GroupMemberModel
	if err := t.db.DbConn.
		Where("group_id = ? AND role = ?", t.currentGroupID, "member").
		Order("joined_at DESC").
		First(&memberModel).Error; err == nil {
		t.currentMemberID = memberModel.ID
	}
}

// aGoalExistsForCategoryWithLimit creates a goal for the specified category with the given limit amount.
func (t *testContext) aGoalExistsForCategoryWithLimit(categoryName, limitAmount string) error {
	// Find the category by name
	var categoryModel model.CategoryModel
	if err := t.db.DbConn.Where("name = ? AND owner_id = ?", categoryName, t.currentUserID).First(&categoryModel).Error; err != nil {
		return fmt.Errorf("category '%s' not found: %w", categoryName, err)
	}

	// Parse limit amount
	limit, err := strconv.ParseFloat(limitAmount, 64)
	if err != nil {
		return fmt.Errorf("invalid limit amount '%s': %w", limitAmount, err)
	}

	goalID := uuid.New()
	t.currentGoalID = goalID

	now := time.Now().UTC()
	goalModel := &model.GoalModel{
		ID:            goalID,
		UserID:        t.currentUserID,
		CategoryID:    categoryModel.ID,
		LimitAmount:   limit,
		AlertOnExceed: true,
		Period:        "monthly",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result := t.db.DbConn.Create(goalModel)
	return result.Error
}

// expenseTransactionsExistForCategoryTrendsTesting creates sample expense transactions for category trends testing.
func (t *testContext) expenseTransactionsExistForCategoryTrendsTesting() error {
	// Get existing categories for the user (created in previous steps)
	var categories []model.CategoryModel
	if err := t.db.DbConn.Where("owner_id = ? AND type = ?", t.currentUserID, "expense").Find(&categories).Error; err != nil {
		return fmt.Errorf("failed to find categories: %w", err)
	}

	// If no categories exist, create one
	if len(categories) == 0 {
		categoryID := uuid.New()
		t.currentCategoryID = categoryID
		categoryModel := &model.CategoryModel{
			ID:        categoryID,
			Name:      "Test Category",
			Color:     "#6366F1",
			Icon:      "tag",
			OwnerType: "user",
			OwnerID:   t.currentUserID,
			Type:      "expense",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := t.db.DbConn.Create(categoryModel).Error; err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}
		categories = append(categories, *categoryModel)
	}

	now := time.Now().UTC()

	// Create expense transactions across different dates
	testDates := []time.Time{
		time.Date(2024, 11, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 11, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 11, 3, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 11, 8, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 11, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 11, 22, 12, 0, 0, 0, time.UTC),
	}

	for i, date := range testDates {
		// Use the first category available, or alternate if multiple exist
		categoryIdx := i % len(categories)
		categoryID := categories[categoryIdx].ID

		transactionModel := &model.TransactionModel{
			ID:          uuid.New(),
			UserID:      t.currentUserID,
			Date:        date,
			Description: fmt.Sprintf("Test expense %d", i+1),
			Amount:      decimal.NewFromFloat(-100.00 * float64(i+1)), // -100, -200, -300, etc.
			Type:        "expense",
			CategoryID:  &categoryID,
			Notes:       "",
			IsRecurring: false,
			IsHidden:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := t.db.DbConn.Create(transactionModel).Error; err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
	}

	return nil
}

// ============================================================================
// Email Notification Step Definitions
// ============================================================================

// mockEmailSender is a mock implementation of adapter.EmailSender for testing.
type mockEmailSender struct {
	shouldFailTemp bool
	shouldFailPerm bool
	lastResendID   string
}

func newMockEmailSender() *mockEmailSender {
	return &mockEmailSender{}
}

func (m *mockEmailSender) Send(ctx context.Context, input adapter.SendEmailInput) (*adapter.SendEmailResult, error) {
	if m.shouldFailPerm {
		return nil, domainerror.NewEmailError(
			domainerror.ErrCodePermanentEmailFailure,
			"permanent email failure",
			errors.New("invalid recipient"),
		)
	}
	if m.shouldFailTemp {
		return nil, errors.New("temporary email failure")
	}
	m.lastResendID = uuid.New().String()
	return &adapter.SendEmailResult{
		ResendID: m.lastResendID,
	}, nil
}

// aPendingEmailJobExistsFor creates a pending email job for testing.
func (t *testContext) aPendingEmailJobExistsFor(email string) error {
	jobID := uuid.New()
	t.lastEmailJobID = jobID

	now := time.Now().UTC()
	templateData := `{"user_name":"Test User","reset_url":"http://localhost:3000/reset?token=test","expires_in":"1 hora"}`

	emailJob := &model.EmailQueueModel{
		ID:             jobID,
		TemplateType:   "password_reset",
		RecipientEmail: email,
		RecipientName:  "Test User",
		Subject:        "Reset your password",
		TemplateData:   templateData,
		Status:         "pending",
		Attempts:       0,
		MaxAttempts:    3,
		CreatedAt:      now,
		ScheduledAt:    now,
	}

	result := t.db.DbConn.Create(emailJob)
	return result.Error
}

// anEmailJobWithFailedAttemptsExistsFor creates an email job with failed attempts.
func (t *testContext) anEmailJobWithFailedAttemptsExistsFor(attempts int, email string) error {
	jobID := uuid.New()
	t.lastEmailJobID = jobID

	now := time.Now().UTC()
	templateData := `{"user_name":"Test User","reset_url":"http://localhost:3000/reset?token=test","expires_in":"1 hora"}`

	emailJob := &model.EmailQueueModel{
		ID:             jobID,
		TemplateType:   "password_reset",
		RecipientEmail: email,
		RecipientName:  "Test User",
		Subject:        "Reset your password",
		TemplateData:   templateData,
		Status:         "pending",
		Attempts:       attempts,
		MaxAttempts:    3,
		LastError:      "previous attempt failed",
		CreatedAt:      now,
		ScheduledAt:    now,
	}

	result := t.db.DbConn.Create(emailJob)
	return result.Error
}

// theEmailSenderWillFailWithATemporaryError sets up the mock to fail with temporary error.
func (t *testContext) theEmailSenderWillFailWithATemporaryError() error {
	if t.emailSenderMock == nil {
		t.emailSenderMock = newMockEmailSender()
	}
	t.emailSenderMock.shouldFailTemp = true
	t.emailSenderMock.shouldFailPerm = false
	return nil
}

// theEmailSenderWillFailWithAPermanentError sets up the mock to fail with permanent error.
func (t *testContext) theEmailSenderWillFailWithAPermanentError() error {
	if t.emailSenderMock == nil {
		t.emailSenderMock = newMockEmailSender()
	}
	t.emailSenderMock.shouldFailTemp = false
	t.emailSenderMock.shouldFailPerm = true
	return nil
}

// theEmailWorkerProcessesTheQueue processes the email queue using a worker.
func (t *testContext) theEmailWorkerProcessesTheQueue() error {
	// Ensure mock sender is initialized
	if t.emailSenderMock == nil {
		t.emailSenderMock = newMockEmailSender()
	}

	// Create email queue repository
	emailQueueRepo := persistence.NewEmailQueueRepository(t.db.DbConn)

	// Create a minimal template renderer for testing
	renderer, err := templates.NewRenderer()
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	// Create worker with test configuration
	config := email.WorkerConfig{
		PollInterval: 1 * time.Second,
		BatchSize:    10,
	}
	worker := email.NewWorker(emailQueueRepo, t.emailSenderMock, renderer, config)

	// Process the queue once
	ctx := context.Background()
	worker.ProcessNow(ctx)

	return nil
}

// anEmailJobShouldBeQueuedFor checks that an email job exists for the given email.
func (t *testContext) anEmailJobShouldBeQueuedFor(email string) error {
	var id string
	result := t.db.DbConn.Raw("SELECT id FROM email_queue WHERE recipient_email = ? ORDER BY created_at DESC LIMIT 1", email).Row()
	if err := result.Scan(&id); err != nil {
		return fmt.Errorf("no email job found for %s: %w", email, err)
	}
	t.lastEmailJobID = uuid.MustParse(id)
	return nil
}

// noEmailJobShouldBeQueuedFor checks that no email job exists for the given email.
func (t *testContext) noEmailJobShouldBeQueuedFor(email string) error {
	var count int64
	result := t.db.DbConn.Raw("SELECT COUNT(*) FROM email_queue WHERE recipient_email = ?", email).Row()
	if err := result.Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("expected no email job for %s, but found %d", email, count)
	}
	return nil
}

// theEmailJobShouldHaveTemplateType checks the template type of the last email job.
func (t *testContext) theEmailJobShouldHaveTemplateType(templateType string) error {
	var actualTemplateType string
	result := t.db.DbConn.Raw("SELECT template_type FROM email_queue WHERE id = ?", t.lastEmailJobID).Row()
	if err := result.Scan(&actualTemplateType); err != nil {
		return fmt.Errorf("email job not found: %w", err)
	}
	if actualTemplateType != templateType {
		return fmt.Errorf("expected template type %s, got %s", templateType, actualTemplateType)
	}
	return nil
}

// theEmailJobShouldHaveStatus checks the status of the last email job.
func (t *testContext) theEmailJobShouldHaveStatus(status string) error {
	var actualStatus string
	result := t.db.DbConn.Raw("SELECT status FROM email_queue WHERE id = ?", t.lastEmailJobID).Row()
	if err := result.Scan(&actualStatus); err != nil {
		return fmt.Errorf("email job not found: %w", err)
	}
	if actualStatus != status {
		return fmt.Errorf("expected status %s, got %s", status, actualStatus)
	}
	return nil
}

// theEmailJobForShouldHaveStatus checks the status of the email job for a specific recipient.
func (t *testContext) theEmailJobForShouldHaveStatus(email, status string) error {
	var id string
	var actualStatus string
	result := t.db.DbConn.Raw("SELECT id, status FROM email_queue WHERE recipient_email = ? ORDER BY created_at DESC LIMIT 1", email).Row()
	if err := result.Scan(&id, &actualStatus); err != nil {
		return fmt.Errorf("email job not found for %s: %w", email, err)
	}
	t.lastEmailJobID = uuid.MustParse(id)
	if actualStatus != status {
		return fmt.Errorf("expected status %s, got %s", status, actualStatus)
	}
	return nil
}

// theEmailJobShouldHaveAResendId checks that the email job has a resend_id.
func (t *testContext) theEmailJobShouldHaveAResendId() error {
	var resendID string
	result := t.db.DbConn.Raw("SELECT resend_id FROM email_queue WHERE id = ?", t.lastEmailJobID).Row()
	if err := result.Scan(&resendID); err != nil {
		return fmt.Errorf("email job not found: %w", err)
	}
	if resendID == "" {
		return fmt.Errorf("expected resend_id to be set, but it was empty")
	}
	return nil
}

// theEmailJobShouldHaveAttemptsEqualTo checks the attempts count of the email job.
func (t *testContext) theEmailJobShouldHaveAttemptsEqualTo(expected int) error {
	var attempts int
	result := t.db.DbConn.Raw("SELECT attempts FROM email_queue WHERE id = ?", t.lastEmailJobID).Row()
	if err := result.Scan(&attempts); err != nil {
		return fmt.Errorf("email job not found: %w", err)
	}
	if attempts != expected {
		return fmt.Errorf("expected attempts %d, got %d", expected, attempts)
	}
	return nil
}

// theEmailJobShouldHaveScheduledAtInTheFuture checks that scheduled_at is in the future.
func (t *testContext) theEmailJobShouldHaveScheduledAtInTheFuture() error {
	var scheduledAt time.Time
	result := t.db.DbConn.Raw("SELECT scheduled_at FROM email_queue WHERE id = ?", t.lastEmailJobID).Row()
	if err := result.Scan(&scheduledAt); err != nil {
		return fmt.Errorf("email job not found: %w", err)
	}
	if !scheduledAt.After(time.Now().UTC()) {
		return fmt.Errorf("expected scheduled_at to be in the future, got %v", scheduledAt)
	}
	return nil
}

// theEmailJobShouldHaveALastError checks that the email job has a last_error.
func (t *testContext) theEmailJobShouldHaveALastError() error {
	var lastError string
	result := t.db.DbConn.Raw("SELECT last_error FROM email_queue WHERE id = ?", t.lastEmailJobID).Row()
	if err := result.Scan(&lastError); err != nil {
		return fmt.Errorf("email job not found: %w", err)
	}
	if lastError == "" {
		return fmt.Errorf("expected last_error to be set, but it was empty")
	}
	return nil
}

// theEmailJobTemplateDataShouldContain checks that template data contains a key.
func (t *testContext) theEmailJobTemplateDataShouldContain(key string) error {
	var templateDataStr string
	result := t.db.DbConn.Raw("SELECT template_data FROM email_queue WHERE id = ?", t.lastEmailJobID).Row()
	if err := result.Scan(&templateDataStr); err != nil {
		return fmt.Errorf("email job not found: %w", err)
	}

	var templateData map[string]interface{}
	if err := json.Unmarshal([]byte(templateDataStr), &templateData); err != nil {
		return fmt.Errorf("failed to parse template data: %w", err)
	}

	if _, ok := templateData[key]; !ok {
		return fmt.Errorf("expected template data to contain key '%s', got %v", key, templateData)
	}
	return nil
}

// theDatabaseShouldContainAnEmailQueueRecordWith checks email_queue record with specific values.
func (t *testContext) theDatabaseShouldContainAnEmailQueueRecordWith(table *godog.Table) error {
	criteria := make(map[string]interface{})
	for _, row := range table.Rows {
		if len(row.Cells) >= 2 {
			key := row.Cells[0].Value
			value := row.Cells[1].Value

			// Convert string values to appropriate types
			switch key {
			case "attempts", "max_attempts":
				intVal, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("invalid integer value for %s: %w", key, err)
				}
				criteria[key] = intVal
			default:
				criteria[key] = value
			}
		}
	}

	// Build the WHERE clause dynamically
	var conditions []string
	var args []interface{}
	for key, value := range criteria {
		conditions = append(conditions, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	whereClause := strings.Join(conditions, " AND ")
	query := fmt.Sprintf("SELECT id FROM email_queue WHERE %s LIMIT 1", whereClause)

	var id string
	result := t.db.DbConn.Raw(query, args...).Row()
	if err := result.Scan(&id); err != nil {
		return fmt.Errorf("no email_queue record found matching criteria %v: %w", criteria, err)
	}

	t.lastEmailJobID = uuid.MustParse(id)
	return nil
}

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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/finance-tracker/backend/internal/application/usecase/auth"
	"github.com/finance-tracker/backend/internal/application/usecase/category"
	categoryrule "github.com/finance-tracker/backend/internal/application/usecase/category_rule"
	"github.com/finance-tracker/backend/internal/application/usecase/goal"
	"github.com/finance-tracker/backend/internal/application/usecase/group"
	"github.com/finance-tracker/backend/internal/application/usecase/transaction"
	"github.com/finance-tracker/backend/internal/infra/server/router"
	"github.com/finance-tracker/backend/internal/integration/adapters"
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
	uri               string
	headers           map[string]string
	client            *http.Client
	response          *response
	db                *mock.Db
	timeMock          *mock.Time
	serverPort        int
	accessToken       string
	refreshToken      string
	resetToken        string
	expiredToken      string
	currentUserID     uuid.UUID
	currentCategoryID uuid.UUID
	currentGoalID     uuid.UUID
	currentGroupID    uuid.UUID
	currentMemberID   uuid.UUID
	currentInviteToken string
	transactionIDs    []uuid.UUID
	lastTransactionID uuid.UUID
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

			// Create adapters/services
			passwordService := adapters.NewPasswordService()
			tokenService := adapters.NewTokenService("test-jwt-secret-key-for-testing-purposes", tokenRepo)
			resetTokenService := adapters.NewPasswordResetTokenService(tokenRepo)

			// Create auth use cases
			registerUseCase := auth.NewRegisterUserUseCase(userRepo, passwordService, tokenService)
			loginUseCase := auth.NewLoginUserUseCase(userRepo, passwordService, tokenService)
			refreshTokenUseCase := auth.NewRefreshTokenUseCase(tokenService)
			logoutUseCase := auth.NewLogoutUserUseCase(tokenService)
			forgotPasswordUseCase := auth.NewForgotPasswordUseCase(userRepo, resetTokenService)
			resetPasswordUseCase := auth.NewResetPasswordUseCase(userRepo, passwordService, resetTokenService)

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

			// Create group use cases
			createGroupUseCase := group.NewCreateGroupUseCase(groupRepo, userRepo)
			listGroupsUseCase := group.NewListGroupsUseCase(groupRepo)
			getGroupUseCase := group.NewGetGroupUseCase(groupRepo)
			inviteMemberUseCase := group.NewInviteMemberUseCase(groupRepo, userRepo)
			acceptInviteUseCase := group.NewAcceptInviteUseCase(groupRepo, userRepo)
			changeMemberRoleUseCase := group.NewChangeMemberRoleUseCase(groupRepo)
			removeMemberUseCase := group.NewRemoveMemberUseCase(groupRepo)
			leaveGroupUseCase := group.NewLeaveGroupUseCase(groupRepo)
			getGroupDashboardUseCase := group.NewGetGroupDashboardUseCase(groupRepo)

			// Create category rule use cases
			categoryRuleRepo := persistence.NewCategoryRuleRepository(testDB.DbConn)
			listCategoryRulesUseCase := categoryrule.NewListCategoryRulesUseCase(categoryRuleRepo)
			createCategoryRuleUseCase := categoryrule.NewCreateCategoryRuleUseCase(categoryRuleRepo, categoryRepo)
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

			// Create middleware
			loginRateLimiter := middleware.NewRateLimiter()
			authMiddleware := middleware.NewAuthMiddleware(tokenService)

			r := router.NewRouter(healthController, authController, userController, categoryController, transactionController, goalController, groupController, categoryRuleController, loginRateLimiter, authMiddleware)
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

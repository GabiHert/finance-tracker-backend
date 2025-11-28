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
	uri           string
	headers       map[string]string
	client        *http.Client
	response      *response
	db            *mock.Db
	timeMock      *mock.Time
	serverPort    int
	accessToken   string
	refreshToken  string
	resetToken    string
	expiredToken  string
	currentUserID uuid.UUID
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

			// Create middleware
			loginRateLimiter := middleware.NewRateLimiter()
			authMiddleware := middleware.NewAuthMiddleware(tokenService)

			r := router.NewRouter(healthController, authController, categoryController, loginRateLimiter, authMiddleware)
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
	return t.executeRequest(method, path, nil)
}

func (t *testContext) iSendARequestToWithBody(method, path string, body *godog.DocString) error {
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

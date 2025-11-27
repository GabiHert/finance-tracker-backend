// Package steps provides step definitions for BDD integration tests.
package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"

	"github.com/finance-tracker/backend/config"
	"github.com/finance-tracker/backend/internal/infra/server/router"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/controller"
)

// TestContext holds the test state for each scenario.
type TestContext struct {
	// HTTP
	server       *httptest.Server
	engine       *gin.Engine
	response     *http.Response
	responseBody []byte

	// Request building
	requestHeaders map[string]string
	requestBody    interface{}

	// Auth
	accessToken  string
	refreshToken string

	// Config
	cfg *config.Config
}

// contextKey is used to store TestContext in context.Context.
type contextKey struct{}

// GetTestContext retrieves the TestContext from context.
func GetTestContext(ctx context.Context) *TestContext {
	if tc, ok := ctx.Value(contextKey{}).(*TestContext); ok {
		return tc
	}
	return nil
}

// SetTestContext stores the TestContext in context.
func SetTestContext(ctx context.Context, tc *TestContext) context.Context {
	return context.WithValue(ctx, contextKey{}, tc)
}

// InitializeTestSuite sets up resources before any scenarios run.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		// Set Gin to test mode
		gin.SetMode(gin.TestMode)
	})

	ctx.AfterSuite(func() {
		// Cleanup any global resources
	})
}

// InitializeScenario registers all step definitions.
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		tc := &TestContext{
			requestHeaders: make(map[string]string),
			cfg:            config.Load(),
		}

		// Setup test server with health controller
		healthController := controller.NewHealthController(func() bool {
			return false // Mock DB as disconnected for tests
		})
		r := router.NewRouter(healthController)
		tc.engine = r.Setup("test")
		tc.server = httptest.NewServer(tc.engine)

		return SetTestContext(ctx, tc), nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		tc := GetTestContext(ctx)
		if tc != nil && tc.server != nil {
			tc.server.Close()
		}
		return ctx, nil
	})

	// Register step definitions
	registerAPISteps(ctx)
	registerResponseSteps(ctx)
}

// registerAPISteps registers HTTP request steps.
func registerAPISteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the API server is running$`, theAPIServerIsRunning)
	ctx.Step(`^I send a "([^"]*)" request to "([^"]*)"$`, iSendARequestTo)
	ctx.Step(`^I send a "([^"]*)" request to "([^"]*)" with body:$`, iSendARequestToWithBody)
	ctx.Step(`^I set header "([^"]*)" to "([^"]*)"$`, iSetHeaderTo)
	ctx.Step(`^I am authenticated$`, iAmAuthenticated)
	ctx.Step(`^I am authenticated as "([^"]*)"$`, iAmAuthenticatedAs)
}

// registerResponseSteps registers response validation steps.
func registerResponseSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the response status should be (\d+)$`, theResponseStatusShouldBe)
	ctx.Step(`^the response should be JSON$`, theResponseShouldBeJSON)
	ctx.Step(`^the response should contain "([^"]*)"$`, theResponseShouldContain)
	ctx.Step(`^the response field "([^"]*)" should be "([^"]*)"$`, theResponseFieldShouldBe)
	ctx.Step(`^the response field "([^"]*)" should exist$`, theResponseFieldShouldExist)
	ctx.Step(`^the response should match json:$`, theResponseShouldMatchJSON)
}

// Step implementations

func theAPIServerIsRunning(ctx context.Context) error {
	tc := GetTestContext(ctx)
	if tc == nil || tc.server == nil {
		return fmt.Errorf("test server is not running")
	}
	return nil
}

func iSendARequestTo(ctx context.Context, method, endpoint string) (context.Context, error) {
	tc := GetTestContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	url := tc.server.URL + endpoint
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range tc.requestHeaders {
		req.Header.Set(key, value)
	}

	// Add auth token if present
	if tc.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+tc.accessToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to send request: %w", err)
	}

	tc.response = resp
	tc.responseBody, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return ctx, fmt.Errorf("failed to read response body: %w", err)
	}

	return SetTestContext(ctx, tc), nil
}

func iSendARequestToWithBody(ctx context.Context, method, endpoint string, body *godog.DocString) (context.Context, error) {
	tc := GetTestContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	url := tc.server.URL + endpoint
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body.Content))
	if err != nil {
		return ctx, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add headers
	for key, value := range tc.requestHeaders {
		req.Header.Set(key, value)
	}

	// Add auth token if present
	if tc.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+tc.accessToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to send request: %w", err)
	}

	tc.response = resp
	tc.responseBody, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return ctx, fmt.Errorf("failed to read response body: %w", err)
	}

	return SetTestContext(ctx, tc), nil
}

func iSetHeaderTo(ctx context.Context, header, value string) (context.Context, error) {
	tc := GetTestContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	tc.requestHeaders[header] = value
	return SetTestContext(ctx, tc), nil
}

func iAmAuthenticated(ctx context.Context) (context.Context, error) {
	tc := GetTestContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	// Set a mock token for testing
	tc.accessToken = "test-token"
	return SetTestContext(ctx, tc), nil
}

func iAmAuthenticatedAs(ctx context.Context, email string) (context.Context, error) {
	tc := GetTestContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	// Set a mock token for testing
	tc.accessToken = "test-token-" + email
	return SetTestContext(ctx, tc), nil
}

func theResponseStatusShouldBe(ctx context.Context, expectedStatus int) error {
	tc := GetTestContext(ctx)
	if tc == nil {
		return fmt.Errorf("test context not found")
	}
	if tc.response == nil {
		return fmt.Errorf("no response received")
	}
	if tc.response.StatusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d. Body: %s", expectedStatus, tc.response.StatusCode, string(tc.responseBody))
	}
	return nil
}

func theResponseShouldBeJSON(ctx context.Context) error {
	tc := GetTestContext(ctx)
	if tc == nil {
		return fmt.Errorf("test context not found")
	}
	var js json.RawMessage
	if err := json.Unmarshal(tc.responseBody, &js); err != nil {
		return fmt.Errorf("response is not valid JSON: %w", err)
	}
	return nil
}

func theResponseShouldContain(ctx context.Context, expected string) error {
	tc := GetTestContext(ctx)
	if tc == nil {
		return fmt.Errorf("test context not found")
	}
	if !strings.Contains(string(tc.responseBody), expected) {
		return fmt.Errorf("response does not contain '%s'. Body: %s", expected, string(tc.responseBody))
	}
	return nil
}

func theResponseFieldShouldBe(ctx context.Context, field, expected string) error {
	tc := GetTestContext(ctx)
	if tc == nil {
		return fmt.Errorf("test context not found")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(tc.responseBody, &data); err != nil {
		return fmt.Errorf("failed to parse response JSON: %w", err)
	}

	value, ok := data[field]
	if !ok {
		return fmt.Errorf("field '%s' not found in response", field)
	}

	actual := fmt.Sprintf("%v", value)
	if actual != expected {
		return fmt.Errorf("field '%s' expected '%s', got '%s'", field, expected, actual)
	}

	return nil
}

func theResponseFieldShouldExist(ctx context.Context, field string) error {
	tc := GetTestContext(ctx)
	if tc == nil {
		return fmt.Errorf("test context not found")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(tc.responseBody, &data); err != nil {
		return fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if _, ok := data[field]; !ok {
		return fmt.Errorf("field '%s' not found in response", field)
	}

	return nil
}

func theResponseShouldMatchJSON(ctx context.Context, body *godog.DocString) error {
	tc := GetTestContext(ctx)
	if tc == nil {
		return fmt.Errorf("test context not found")
	}

	var expected, actual interface{}

	if err := json.Unmarshal([]byte(body.Content), &expected); err != nil {
		return fmt.Errorf("failed to parse expected JSON: %w", err)
	}

	if err := json.Unmarshal(tc.responseBody, &actual); err != nil {
		return fmt.Errorf("failed to parse response JSON: %w", err)
	}

	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)

	if string(expectedJSON) != string(actualJSON) {
		return fmt.Errorf("expected JSON:\n%s\nactual JSON:\n%s", string(expectedJSON), string(actualJSON))
	}

	return nil
}

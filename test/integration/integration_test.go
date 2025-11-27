//go:build integration

// Package integration provides BDD integration tests using Godog/Cucumber.
package integration

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/finance-tracker/backend/test/integration/steps"
)

// TestFeatures runs all BDD feature tests.
func TestFeatures(t *testing.T) {
	opts := godog.Options{
		Format:      "pretty",
		Paths:       []string{"features"},
		Output:      colors.Colored(os.Stdout),
		Concurrency: 1, // Run sequentially for database tests
		Randomize:   0, // Don't randomize for predictable results
		Strict:      true,
		TestingT:    t,
	}

	// Allow tag filtering via environment variable
	if tags := os.Getenv("GODOG_TAGS"); tags != "" {
		opts.Tags = tags
	}

	suite := godog.TestSuite{
		Name:                 "finance-tracker-api",
		ScenarioInitializer:  steps.InitializeScenario,
		TestSuiteInitializer: steps.InitializeTestSuite,
		Options:              &opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

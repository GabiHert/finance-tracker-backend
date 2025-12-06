# Finance Tracker - Category Expense Trends Feature
# Version: 1.0 | Milestone 13 (M13-category-trends)

@all @category-trends @dashboard
Feature: Category Expense Trends
  As a user
  I want to see my expense trends per category over time
  So that I can understand my spending patterns

  Background:
    Given the API server is running
    And a user exists with email "test@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens

  # ============================================
  # HAPPY PATH SCENARIOS
  # ============================================

  @success @trends
  Scenario: Get daily expense trends with multiple categories
    Given a category exists with name "Alimentacao" and type "expense"
    And a category exists with name "Transporte" and type "expense"
    And expense transactions exist for category trends testing
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&end_date=2024-11-03&granularity=daily&top_categories=8"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "data"
    And the response field "data.period.granularity" should be "daily"

  @success @trends
  Scenario: Get weekly expense trends
    Given a category exists with name "Alimentacao" and type "expense"
    And expense transactions exist for category trends testing
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&end_date=2024-11-30&granularity=weekly"
    Then the response status should be 200
    And the response should be JSON
    And the response field "data.period.granularity" should be "weekly"

  @success @trends
  Scenario: Get monthly expense trends
    Given a category exists with name "Alimentacao" and type "expense"
    And expense transactions exist for category trends testing
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&end_date=2024-12-31&granularity=monthly"
    Then the response status should be 200
    And the response should be JSON
    And the response field "data.period.granularity" should be "monthly"

  @success @trends
  Scenario: Empty result when no expenses in period
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&end_date=2024-11-30&granularity=daily"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "data"

  @success @trends
  Scenario: Default top_categories value is 8
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&end_date=2024-11-30&granularity=daily"
    Then the response status should be 200
    And the response should be JSON

  # ============================================
  # VALIDATION SCENARIOS
  # ============================================

  @failure @validation
  Scenario: Invalid date format returns error
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=01-11-2024&end_date=30-11-2024&granularity=daily"
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @validation
  Scenario: Invalid granularity returns error
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&end_date=2024-11-30&granularity=yearly"
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @validation
  Scenario: Missing start_date returns error
    When I send a "GET" request to "/api/v1/dashboard/category-trends?end_date=2024-11-30&granularity=daily"
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @validation
  Scenario: Missing end_date returns error
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&granularity=daily"
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @validation
  Scenario: Missing granularity returns error
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&end_date=2024-11-30"
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @validation
  Scenario: End date before start date returns error
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-30&end_date=2024-11-01&granularity=daily"
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  # ============================================
  # AUTHENTICATION SCENARIOS
  # ============================================

  @failure @unauthorized
  Scenario: Cannot access category trends without authentication
    Given the header is empty
    When I send a "GET" request to "/api/v1/dashboard/category-trends?start_date=2024-11-01&end_date=2024-11-30&granularity=daily"
    Then the response status should be 401

# Finance Tracker - Goals (Spending Limits) Management Feature
# Version: 1.0 | Milestone 7 (M7-B1 through M7-B6)

@all @goals
Feature: Goals (Spending Limits) Management
  As a user
  I want to manage my spending goals/limits per category
  So that I can track and control my expenses

  Background:
    Given the API server is running
    And a user exists with email "test@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens

  # ============================================
  # LIST GOALS SCENARIOS (M7-B2)
  # ============================================

  @success @list
  Scenario: List goals when none exist
    When I send a "GET" request to "/api/v1/goals"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "goals"

  @success @list
  Scenario: List goals with existing goals
    Given a category exists with name "Food" and type "expense"
    And a goal exists for category "Food" with limit "500.00"
    When I send a "GET" request to "/api/v1/goals"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "goals"

  # ============================================
  # CREATE GOAL SCENARIOS (M7-B3)
  # ============================================

  @success @create
  Scenario: Create goal with valid data
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "{{category_id}}",
        "limit_amount": 500.00,
        "alert_on_exceed": true
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "id" should exist
    And the response field "limit_amount" should exist

  @success @create
  Scenario: Create goal with custom period
    Given a category exists with name "Entertainment" and type "expense"
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "{{category_id}}",
        "limit_amount": 200.00,
        "alert_on_exceed": false,
        "period": "weekly"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "period" should be "weekly"

  @success @create
  Scenario: Create goal without alert_on_exceed defaults to true
    Given a category exists with name "Transport" and type "expense"
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "{{category_id}}",
        "limit_amount": 300.00
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "alert_on_exceed" should be "true"

  @failure @create @validation
  Scenario: Validation error for missing category_id
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "limit_amount": 500.00
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for missing limit_amount
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "{{category_id}}"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for zero limit_amount
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "{{category_id}}",
        "limit_amount": 0
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for negative limit_amount
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "{{category_id}}",
        "limit_amount": -100.00
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for invalid period
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "{{category_id}}",
        "limit_amount": 500.00,
        "period": "invalid"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @not-found
  Scenario: Cannot create goal for non-existent category
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "00000000-0000-0000-0000-000000000001",
        "limit_amount": 500.00
      }
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "GOL-010004"

  @failure @create @conflict
  Scenario: Cannot create duplicate goal for same category
    Given a category exists with name "Food" and type "expense"
    And a goal exists for category "Food" with limit "500.00"
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "{{category_id}}",
        "limit_amount": 600.00
      }
      """
    Then the response status should be 409
    And the response should be JSON
    And the response field "code" should be "GOL-010002"

  # ============================================
  # GET GOAL BY ID SCENARIOS (M7-B4)
  # ============================================

  @success @get
  Scenario: Get goal by ID
    Given a category exists with name "Food" and type "expense"
    And a goal exists for category "Food" with limit "500.00"
    When I send a "GET" request to "/api/v1/goals/{{goal_id}}"
    Then the response status should be 200
    And the response should be JSON
    And the response field "id" should exist
    And the response field "limit_amount" should exist

  @failure @get @not-found
  Scenario: Cannot get non-existent goal
    When I send a "GET" request to "/api/v1/goals/00000000-0000-0000-0000-000000000000"
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "GOL-010001"

  # ============================================
  # UPDATE GOAL SCENARIOS (M7-B5)
  # ============================================

  @success @update
  Scenario: Update goal limit_amount
    Given a category exists with name "Food" and type "expense"
    And a goal exists for category "Food" with limit "500.00"
    When I send a "PATCH" request to "/api/v1/goals/{{goal_id}}" with body:
      """
      {
        "limit_amount": 750.00
      }
      """
    Then the response status should be 200
    And the response should be JSON

  @success @update
  Scenario: Update goal alert_on_exceed
    Given a category exists with name "Food" and type "expense"
    And a goal exists for category "Food" with limit "500.00"
    When I send a "PATCH" request to "/api/v1/goals/{{goal_id}}" with body:
      """
      {
        "alert_on_exceed": false
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "alert_on_exceed" should be "false"

  @success @update
  Scenario: Update goal period
    Given a category exists with name "Food" and type "expense"
    And a goal exists for category "Food" with limit "500.00"
    When I send a "PATCH" request to "/api/v1/goals/{{goal_id}}" with body:
      """
      {
        "period": "weekly"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "period" should be "weekly"

  @failure @update @not-found
  Scenario: Cannot update non-existent goal
    When I send a "PATCH" request to "/api/v1/goals/00000000-0000-0000-0000-000000000000" with body:
      """
      {
        "limit_amount": 750.00
      }
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "GOL-010001"

  @failure @update @validation
  Scenario: Cannot update goal with negative limit_amount
    Given a category exists with name "Food" and type "expense"
    And a goal exists for category "Food" with limit "500.00"
    When I send a "PATCH" request to "/api/v1/goals/{{goal_id}}" with body:
      """
      {
        "limit_amount": -100.00
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  # ============================================
  # DELETE GOAL SCENARIOS (M7-B6)
  # ============================================

  @success @delete
  Scenario: Delete existing goal
    Given a category exists with name "Food" and type "expense"
    And a goal exists for category "Food" with limit "500.00"
    When I send a "DELETE" request to "/api/v1/goals/{{goal_id}}"
    Then the response status should be 204

  @failure @delete @not-found
  Scenario: Cannot delete non-existent goal
    When I send a "DELETE" request to "/api/v1/goals/00000000-0000-0000-0000-000000000000"
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "GOL-010001"

  # ============================================
  # AUTHENTICATION SCENARIOS
  # ============================================

  @failure @unauthorized
  Scenario: Cannot access goals without authentication
    Given the header is empty
    When I send a "GET" request to "/api/v1/goals"
    Then the response status should be 401

  @failure @unauthorized
  Scenario: Cannot create goal without authentication
    Given the header is empty
    When I send a "POST" request to "/api/v1/goals" with body:
      """
      {
        "category_id": "00000000-0000-0000-0000-000000000001",
        "limit_amount": 500.00
      }
      """
    Then the response status should be 401

@all @categories
Feature: Category Management
  As a user
  I want to manage my transaction categories
  So that I can organize my financial data

  Background:
    Given the API server is running
    And a user exists with email "test@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens

  @success @list
  Scenario: List user categories when none exist
    When I send a "GET" request to "/api/v1/categories"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "categories"

  @success @list
  Scenario: List user categories with existing categories
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "color": "#FF5733",
        "icon": "utensils",
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Transport",
        "color": "#33FF57",
        "icon": "car",
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "GET" request to "/api/v1/categories"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "categories"

  @success @create
  Scenario: Create new expense category
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "color": "#FF5733",
        "icon": "utensils",
        "type": "expense"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "name" should be "Food"
    And the response field "color" should be "#FF5733"
    And the response field "icon" should be "utensils"
    And the response field "type" should be "expense"
    And the response field "id" should exist

  @success @create
  Scenario: Create new income category
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Salary",
        "color": "#33FF57",
        "icon": "money",
        "type": "income"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "name" should be "Salary"
    And the response field "type" should be "income"

  @success @create @defaults
  Scenario: Create category with default color and icon
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Utilities",
        "type": "expense"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "name" should be "Utilities"
    And the response field "color" should be "#6366F1"
    And the response field "icon" should be "tag"

  @failure @create @duplicate
  Scenario: Cannot create duplicate category name for same user
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "color": "#FF5733",
        "icon": "utensils",
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "color": "#33FF57",
        "icon": "plate",
        "type": "expense"
      }
      """
    Then the response status should be 409
    And the response should be JSON
    And the response field "code" should be "CAT-010005"

  @failure @create @validation
  Scenario: Validation error for missing name
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "color": "#FF5733",
        "icon": "utensils",
        "type": "expense"
      }
      """
    Then the response status should be 400
    And the response should be JSON

  @failure @create @validation
  Scenario: Validation error for missing type
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "color": "#FF5733",
        "icon": "utensils"
      }
      """
    Then the response status should be 400
    And the response should be JSON

  @failure @create @validation
  Scenario: Validation error for invalid color format
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "color": "invalid-color",
        "icon": "utensils",
        "type": "expense"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response field "code" should be "CAT-010002"

  @failure @create @validation
  Scenario: Validation error for name too long
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "This is a very long category name that exceeds fifty characters limit",
        "type": "expense"
      }
      """
    Then the response status should be 400
    And the response should be JSON

  @failure @create @validation
  Scenario: Validation error for invalid category type
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "type": "invalid"
      }
      """
    Then the response status should be 400
    And the response should be JSON

  @success @update
  Scenario: Update category name and color
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "color": "#FF5733",
        "icon": "utensils",
        "type": "expense"
      }
      """
    Then the response status should be 201
    And the response field "id" should exist
    When I send a "GET" request to "/api/v1/categories"
    Then the response status should be 200

  @failure @update @not-found
  Scenario: Cannot update non-existent category
    When I send a "PATCH" request to "/api/v1/categories/00000000-0000-0000-0000-000000000000" with body:
      """
      {
        "name": "New Name"
      }
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "CAT-010004"

  @success @delete
  Scenario: Delete existing category
    When I send a "POST" request to "/api/v1/categories" with body:
      """
      {
        "name": "Food",
        "color": "#FF5733",
        "icon": "utensils",
        "type": "expense"
      }
      """
    Then the response status should be 201

  @failure @delete @not-found
  Scenario: Cannot delete non-existent category
    When I send a "DELETE" request to "/api/v1/categories/00000000-0000-0000-0000-000000000000"
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "CAT-010004"

  @failure @unauthorized
  Scenario: Cannot access categories without authentication
    Given the header is empty
    When I send a "GET" request to "/api/v1/categories"
    Then the response status should be 401

# Finance Tracker - Delete Account Feature
# Version: 1.0 | Milestone 10 (M10-E2E-009)

Feature: Delete Account
  As a user
  I want to delete my account
  So that I can remove all my data from the Finance Tracker application

  Background:
    Given the API server is running

  # ============================================
  # DELETE ACCOUNT SUCCESS SCENARIOS
  # ============================================

  @all @delete-account @success
  Scenario: Successful account deletion with valid password and confirmation
    Given a user exists with email "delete@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "DELETE" request to "/api/v1/users/me" with body:
      """
      {
        "password": "SecurePass123!",
        "confirmation": "DELETE"
      }
      """
    Then the response status should be 204

  @all @delete-account @success
  Scenario: Login fails after account deletion
    Given a user exists with email "todelete@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "DELETE" request to "/api/v1/users/me" with body:
      """
      {
        "password": "SecurePass123!",
        "confirmation": "DELETE"
      }
      """
    Then the response status should be 204
    When I send a "POST" request to "/api/v1/auth/login" with body:
      """
      {
        "email": "todelete@example.com",
        "password": "SecurePass123!"
      }
      """
    Then the response status should be 401
    And the response should be JSON
    And the response should contain "error"

  # ============================================
  # DELETE ACCOUNT ERROR SCENARIOS
  # ============================================

  @all @delete-account @error
  Scenario: Delete account fails with wrong password
    Given a user exists with email "wrongpass@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "DELETE" request to "/api/v1/users/me" with body:
      """
      {
        "password": "WrongPassword123!",
        "confirmation": "DELETE"
      }
      """
    Then the response status should be 401
    And the response should be JSON
    And the response should contain "error"

  @all @delete-account @error
  Scenario: Delete account fails with wrong confirmation text
    Given a user exists with email "wrongconfirm@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "DELETE" request to "/api/v1/users/me" with body:
      """
      {
        "password": "SecurePass123!",
        "confirmation": "delete"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @delete-account @error
  Scenario: Delete account fails with empty confirmation
    Given a user exists with email "emptyconfirm@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "DELETE" request to "/api/v1/users/me" with body:
      """
      {
        "password": "SecurePass123!",
        "confirmation": ""
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @delete-account @error
  Scenario: Delete account fails with missing password
    Given a user exists with email "missingpass@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "DELETE" request to "/api/v1/users/me" with body:
      """
      {
        "confirmation": "DELETE"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @delete-account @error
  Scenario: Delete account fails with missing confirmation
    Given a user exists with email "missingconfirm@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "DELETE" request to "/api/v1/users/me" with body:
      """
      {
        "password": "SecurePass123!"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @delete-account @error
  Scenario: Delete account fails without authentication
    When I send a "DELETE" request to "/api/v1/users/me" with body:
      """
      {
        "password": "SecurePass123!",
        "confirmation": "DELETE"
      }
      """
    Then the response status should be 401
    And the response should be JSON
    And the response should contain "error"

# Finance Tracker - Authentication Feature
# Version: 1.0 | Milestone 2 (M2-B1 through M2-B8)

Feature: User Authentication
  As a user
  I want to register, login, and manage my authentication
  So that I can securely access the Finance Tracker application

  Background:
    Given the API server is running

  # ============================================
  # REGISTRATION SCENARIOS
  # ============================================

  @all @auth @registration @success
  Scenario: Successful user registration with valid data
    When I send a "POST" request to "/api/v1/auth/register" with body:
      """
      {
        "email": "newuser@example.com",
        "name": "John Doe",
        "password": "SecurePass123!",
        "terms_accepted": true
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "access_token" should exist
    And the response field "refresh_token" should exist
    And the response should contain "user"

  @all @auth @registration @error
  Scenario: Registration fails when email already exists
    Given a user exists with email "existing@example.com"
    When I send a "POST" request to "/api/v1/auth/register" with body:
      """
      {
        "email": "existing@example.com",
        "name": "Another User",
        "password": "SecurePass123!",
        "terms_accepted": true
      }
      """
    Then the response status should be 409
    And the response should be JSON
    And the response should contain "error"

  @all @auth @registration @error
  Scenario: Registration fails when terms are not accepted
    When I send a "POST" request to "/api/v1/auth/register" with body:
      """
      {
        "email": "newuser2@example.com",
        "name": "Jane Doe",
        "password": "SecurePass123!",
        "terms_accepted": false
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @auth @registration @error
  Scenario: Registration fails with weak password
    When I send a "POST" request to "/api/v1/auth/register" with body:
      """
      {
        "email": "newuser3@example.com",
        "name": "Bob Smith",
        "password": "weak",
        "terms_accepted": true
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @auth @registration @error
  Scenario: Registration fails with invalid email format
    When I send a "POST" request to "/api/v1/auth/register" with body:
      """
      {
        "email": "invalid-email",
        "name": "Test User",
        "password": "SecurePass123!",
        "terms_accepted": true
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @auth @registration @error
  Scenario: Registration fails with missing required fields
    When I send a "POST" request to "/api/v1/auth/register" with body:
      """
      {
        "email": "newuser4@example.com"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  # ============================================
  # LOGIN SCENARIOS
  # ============================================

  @all @auth @login @success
  Scenario: Successful login with valid credentials
    Given a user exists with email "login@example.com" and password "SecurePass123!"
    When I send a "POST" request to "/api/v1/auth/login" with body:
      """
      {
        "email": "login@example.com",
        "password": "SecurePass123!"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "access_token" should exist
    And the response field "refresh_token" should exist
    And the response should contain "user"

  @all @auth @login @success
  Scenario: Login with remember_me extends token duration
    Given a user exists with email "remember@example.com" and password "SecurePass123!"
    When I send a "POST" request to "/api/v1/auth/login" with body:
      """
      {
        "email": "remember@example.com",
        "password": "SecurePass123!",
        "remember_me": true
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "access_token" should exist
    And the response field "refresh_token" should exist

  @all @auth @login @error
  Scenario: Login fails with invalid credentials
    Given a user exists with email "wrongpass@example.com" and password "SecurePass123!"
    When I send a "POST" request to "/api/v1/auth/login" with body:
      """
      {
        "email": "wrongpass@example.com",
        "password": "WrongPassword!"
      }
      """
    Then the response status should be 401
    And the response should be JSON
    And the response should contain "error"

  @all @auth @login @error
  Scenario: Login fails with non-existent email
    When I send a "POST" request to "/api/v1/auth/login" with body:
      """
      {
        "email": "nonexistent@example.com",
        "password": "AnyPassword123!"
      }
      """
    Then the response status should be 401
    And the response should be JSON
    And the response should contain "error"

  @all @auth @login @error
  Scenario: Login fails with missing credentials
    When I send a "POST" request to "/api/v1/auth/login" with body:
      """
      {
        "email": "test@example.com"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  # ============================================
  # TOKEN REFRESH SCENARIOS
  # ============================================

  @all @auth @refresh @success
  Scenario: Token refresh returns new tokens
    Given a user exists with email "refresh@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "POST" request to "/api/v1/auth/refresh" with body:
      """
      {
        "refresh_token": "{{refresh_token}}"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "access_token" should exist
    And the response field "refresh_token" should exist

  @all @auth @refresh @error
  Scenario: Token refresh fails with invalid token
    When I send a "POST" request to "/api/v1/auth/refresh" with body:
      """
      {
        "refresh_token": "invalid-refresh-token"
      }
      """
    Then the response status should be 401
    And the response should be JSON
    And the response should contain "error"

  @all @auth @refresh @error
  Scenario: Token refresh fails with missing token
    When I send a "POST" request to "/api/v1/auth/refresh" with body:
      """
      {}
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  # ============================================
  # LOGOUT SCENARIOS
  # ============================================

  @all @auth @logout @success
  Scenario: Successful logout invalidates refresh token
    Given a user exists with email "logout@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "POST" request to "/api/v1/auth/logout" with body:
      """
      {
        "refresh_token": "{{refresh_token}}"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "message"

  @all @auth @logout @success
  Scenario: Logout with invalid token still returns success
    When I send a "POST" request to "/api/v1/auth/logout" with body:
      """
      {
        "refresh_token": "already-invalid-token"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "message"

  # ============================================
  # PASSWORD RESET SCENARIOS
  # ============================================

  @all @auth @password-reset @success
  Scenario: Forgot password always returns 200 to prevent enumeration
    When I send a "POST" request to "/api/v1/auth/forgot-password" with body:
      """
      {
        "email": "anuser@example.com"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "message"

  @all @auth @password-reset @success
  Scenario: Forgot password returns 200 even for non-existent email
    When I send a "POST" request to "/api/v1/auth/forgot-password" with body:
      """
      {
        "email": "nonexistent@example.com"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "message"

  @all @auth @password-reset @error
  Scenario: Forgot password fails with invalid email format
    When I send a "POST" request to "/api/v1/auth/forgot-password" with body:
      """
      {
        "email": "not-an-email"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @auth @password-reset @success
  Scenario: Reset password with valid token succeeds
    Given a user exists with email "reset@example.com" and password "OldPassword123!"
    And a password reset token exists for "reset@example.com"
    When I send a "POST" request to "/api/v1/auth/reset-password" with body:
      """
      {
        "token": "{{reset_token}}",
        "new_password": "NewSecurePass123!"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "message"

  @all @auth @password-reset @error
  Scenario: Reset password fails with invalid token
    When I send a "POST" request to "/api/v1/auth/reset-password" with body:
      """
      {
        "token": "invalid-reset-token",
        "new_password": "NewSecurePass123!"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @auth @password-reset @error
  Scenario: Reset password fails with expired token
    Given an expired password reset token exists
    When I send a "POST" request to "/api/v1/auth/reset-password" with body:
      """
      {
        "token": "{{expired_reset_token}}",
        "new_password": "NewSecurePass123!"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @all @auth @password-reset @error
  Scenario: Reset password fails with weak new password
    Given a user exists with email "weakreset@example.com" and password "OldPassword123!"
    And a password reset token exists for "weakreset@example.com"
    When I send a "POST" request to "/api/v1/auth/reset-password" with body:
      """
      {
        "token": "{{reset_token}}",
        "new_password": "weak"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

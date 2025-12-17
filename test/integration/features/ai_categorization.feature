@all @ai_categorization
Feature: AI Smart Categorization
  As a user
  I want to automatically categorize my uncategorized transactions using AI
  So that I can save time organizing my financial data

  Background:
    Given the API server is running
    And a user exists with email "test@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens

  # ============================================================================
  # GET STATUS SCENARIOS
  # ============================================================================

  @success @status
  Scenario: Get AI categorization status with no uncategorized transactions
    When I send a "GET" request to "/api/v1/ai/categorization/status"
    Then the response status should be 200
    And the response should be JSON
    And the response field "uncategorized_count" should be "0"
    And the response field "is_processing" should be "false"
    And the response field "pending_suggestions_count" should be "0"

  @success @status
  Scenario: Get AI categorization status with uncategorized transactions
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "description": "UBER TRIP",
        "amount": "-25.50",
        "type": "expense",
        "date": "2024-01-15"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "description": "NETFLIX SUBSCRIPTION",
        "amount": "-15.99",
        "type": "expense",
        "date": "2024-01-16"
      }
      """
    Then the response status should be 201
    When I send a "GET" request to "/api/v1/ai/categorization/status"
    Then the response status should be 200
    And the response should be JSON
    And the response field "uncategorized_count" should be "2"

  @failure @unauthorized
  Scenario: Cannot get status without authentication
    Given the header is empty
    When I send a "GET" request to "/api/v1/ai/categorization/status"
    Then the response status should be 401

  # ============================================================================
  # START CATEGORIZATION SCENARIOS
  # ============================================================================

  @success @start
  Scenario: Start AI categorization process
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "description": "UBER TRIP TO AIRPORT",
        "amount": "-45.00",
        "type": "expense",
        "date": "2024-01-15"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/ai/categorization/start" with body:
      """
      {}
      """
    Then the response status should be 202
    And the response should be JSON
    And the response field "job_id" should exist
    And the response field "message" should exist

  @failure @start @no_transactions
  Scenario: Cannot start categorization without uncategorized transactions
    When I send a "POST" request to "/api/v1/ai/categorization/start" with body:
      """
      {}
      """
    Then the response status should be 400
    And the response should be JSON
    And the response field "code" should be "AIC-010003"

  @failure @unauthorized
  Scenario: Cannot start categorization without authentication
    Given the header is empty
    When I send a "POST" request to "/api/v1/ai/categorization/start" with body:
      """
      {}
      """
    Then the response status should be 401

  # ============================================================================
  # GET SUGGESTIONS SCENARIOS
  # ============================================================================

  @success @suggestions
  Scenario: Get pending suggestions when none exist
    When I send a "GET" request to "/api/v1/ai/categorization/suggestions"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "suggestions"

  @failure @unauthorized
  Scenario: Cannot get suggestions without authentication
    Given the header is empty
    When I send a "GET" request to "/api/v1/ai/categorization/suggestions"
    Then the response status should be 401

  # ============================================================================
  # APPROVE SUGGESTION SCENARIOS
  # ============================================================================

  @failure @approve @not_found
  Scenario: Cannot approve non-existent suggestion
    When I send a "POST" request to "/api/v1/ai/categorization/suggestions/00000000-0000-0000-0000-000000000000/approve" with body:
      """
      {}
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "AIC-010001"

  @failure @approve @validation
  Scenario: Cannot approve suggestion with invalid UUID
    When I send a "POST" request to "/api/v1/ai/categorization/suggestions/invalid-uuid/approve" with body:
      """
      {}
      """
    Then the response status should be 400
    And the response should be JSON

  @failure @unauthorized
  Scenario: Cannot approve suggestion without authentication
    Given the header is empty
    When I send a "POST" request to "/api/v1/ai/categorization/suggestions/00000000-0000-0000-0000-000000000000/approve" with body:
      """
      {}
      """
    Then the response status should be 401

  # ============================================================================
  # REJECT SUGGESTION SCENARIOS
  # ============================================================================

  @failure @reject @not_found
  Scenario: Cannot reject non-existent suggestion
    When I send a "POST" request to "/api/v1/ai/categorization/suggestions/00000000-0000-0000-0000-000000000000/reject" with body:
      """
      {
        "action": "skip"
      }
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "AIC-010001"

  @failure @reject @validation
  Scenario: Cannot reject with invalid action
    When I send a "POST" request to "/api/v1/ai/categorization/suggestions/00000000-0000-0000-0000-000000000000/reject" with body:
      """
      {
        "action": "invalid_action"
      }
      """
    Then the response status should be 400
    And the response should be JSON

  @failure @unauthorized
  Scenario: Cannot reject suggestion without authentication
    Given the header is empty
    When I send a "POST" request to "/api/v1/ai/categorization/suggestions/00000000-0000-0000-0000-000000000000/reject" with body:
      """
      {
        "action": "skip"
      }
      """
    Then the response status should be 401

  # ============================================================================
  # CLEAR SUGGESTIONS SCENARIOS
  # ============================================================================

  @success @clear
  Scenario: Clear all pending suggestions when none exist
    When I send a "DELETE" request to "/api/v1/ai/categorization/suggestions"
    Then the response status should be 200
    And the response should be JSON
    And the response field "deleted_count" should be "0"

  @failure @unauthorized
  Scenario: Cannot clear suggestions without authentication
    Given the header is empty
    When I send a "DELETE" request to "/api/v1/ai/categorization/suggestions"
    Then the response status should be 401

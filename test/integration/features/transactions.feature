# Finance Tracker - Transaction Management Feature
# Version: 1.0 | Milestone 4 (M4-B1 through M4-B7)

@all @transactions
Feature: Transaction Management
  As a user
  I want to manage my financial transactions
  So that I can track my income and expenses

  Background:
    Given the API server is running
    And a user exists with email "test@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens

  # ============================================
  # LIST TRANSACTIONS SCENARIOS (M4-B2)
  # ============================================

  @success @list
  Scenario: List transactions when none exist
    When I send a "GET" request to "/api/v1/transactions"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "transactions"
    And the response should contain "pagination"
    And the response should contain "totals"

  @success @list @pagination
  Scenario: List transactions with pagination defaults
    When I send a "GET" request to "/api/v1/transactions"
    Then the response status should be 200
    And the response should be JSON
    And the response field "pagination.page" should be "1"
    And the response field "pagination.limit" should be "20"

  @success @list @filter
  Scenario: Filter transactions by type expense
    When I send a "GET" request to "/api/v1/transactions?type=expense"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "transactions"

  @success @list @filter
  Scenario: Filter transactions by type income
    When I send a "GET" request to "/api/v1/transactions?type=income"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "transactions"

  @success @list @filter
  Scenario: Filter transactions by date range
    When I send a "GET" request to "/api/v1/transactions?startDate=2025-01-01&endDate=2025-12-31"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "transactions"

  @success @list @search
  Scenario: Search transactions by description
    When I send a "GET" request to "/api/v1/transactions?search=grocery"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "transactions"

  # ============================================
  # CREATE TRANSACTION SCENARIOS (M4-B3)
  # ============================================

  @success @create
  Scenario: Create expense transaction with category
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Grocery shopping",
        "amount": -150.50,
        "type": "expense",
        "category_id": "{{category_id}}"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "description" should be "Grocery shopping"
    And the response field "type" should be "expense"
    And the response field "id" should exist

  @success @create
  Scenario: Create income transaction without category
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-15",
        "description": "Salary",
        "amount": 5000.00,
        "type": "income"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "description" should be "Salary"
    And the response field "type" should be "income"
    And the response field "id" should exist

  @success @create
  Scenario: Create transaction with notes
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-18",
        "description": "Coffee",
        "amount": -5.50,
        "type": "expense",
        "notes": "Morning coffee at Starbucks"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "notes" should be "Morning coffee at Starbucks"

  @success @create
  Scenario: Create recurring transaction
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-01",
        "description": "Netflix subscription",
        "amount": -15.99,
        "type": "expense",
        "is_recurring": true
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "is_recurring" should be "true"

  @failure @create @validation
  Scenario: Validation error for missing date
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "description": "Test",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for missing description
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for missing amount
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Test",
        "type": "expense"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for missing type
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Test",
        "amount": -50.00
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for invalid type
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Test",
        "amount": -50.00,
        "type": "invalid"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for invalid date format
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "invalid-date",
        "description": "Test",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @create @validation
  Scenario: Validation error for non-existent category
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Test",
        "amount": -50.00,
        "type": "expense",
        "category_id": "00000000-0000-0000-0000-000000000001"
      }
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "TXN-010006"

  # ============================================
  # UPDATE TRANSACTION SCENARIOS (M4-B4)
  # ============================================

  @success @update
  Scenario: Update transaction description
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Old description",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    And the response field "id" should exist
    When I send a "PATCH" request to "/api/v1/transactions/{{transaction_id}}" with body:
      """
      {
        "description": "New description"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "description" should be "New description"

  @success @update
  Scenario: Update transaction amount
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Test",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "PATCH" request to "/api/v1/transactions/{{transaction_id}}" with body:
      """
      {
        "amount": -75.00
      }
      """
    Then the response status should be 200
    And the response should be JSON

  @success @update
  Scenario: Update transaction category
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Test",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "PATCH" request to "/api/v1/transactions/{{transaction_id}}" with body:
      """
      {
        "category_id": "{{category_id}}"
      }
      """
    Then the response status should be 200
    And the response should be JSON

  @failure @update @not-found
  Scenario: Cannot update non-existent transaction
    When I send a "PATCH" request to "/api/v1/transactions/00000000-0000-0000-0000-000000000000" with body:
      """
      {
        "description": "New description"
      }
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "TXN-010004"

  # ============================================
  # DELETE TRANSACTION SCENARIOS (M4-B5)
  # ============================================

  @success @delete
  Scenario: Delete existing transaction
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "To be deleted",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "DELETE" request to "/api/v1/transactions/{{transaction_id}}"
    Then the response status should be 204

  @failure @delete @not-found
  Scenario: Cannot delete non-existent transaction
    When I send a "DELETE" request to "/api/v1/transactions/00000000-0000-0000-0000-000000000000"
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "TXN-010004"

  # ============================================
  # BULK DELETE SCENARIOS (M4-B6)
  # ============================================

  @success @bulk @delete
  Scenario: Bulk delete transactions
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Transaction 1",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-21",
        "description": "Transaction 2",
        "amount": -60.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/transactions/bulk-delete" with body:
      """
      {
        "ids": {{transaction_ids}}
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "deleted_count" should be "2"

  @failure @bulk @delete @validation
  Scenario: Bulk delete fails with empty ids array
    When I send a "POST" request to "/api/v1/transactions/bulk-delete" with body:
      """
      {
        "ids": []
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @bulk @delete @not-found
  Scenario: Bulk delete fails when transaction not found
    When I send a "POST" request to "/api/v1/transactions/bulk-delete" with body:
      """
      {
        "ids": ["00000000-0000-0000-0000-000000000000"]
      }
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "TXN-010004"

  # ============================================
  # BULK CATEGORIZE SCENARIOS (M4-B7)
  # ============================================

  @success @bulk @categorize
  Scenario: Bulk categorize transactions
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Transaction 1",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-21",
        "description": "Transaction 2",
        "amount": -60.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/transactions/bulk-categorize" with body:
      """
      {
        "ids": {{transaction_ids}},
        "category_id": "{{category_id}}"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "updated_count" should be "2"

  @failure @bulk @categorize @validation
  Scenario: Bulk categorize fails with empty ids array
    Given a category exists with name "Food" and type "expense"
    When I send a "POST" request to "/api/v1/transactions/bulk-categorize" with body:
      """
      {
        "ids": [],
        "category_id": "{{category_id}}"
      }
      """
    Then the response status should be 400
    And the response should be JSON
    And the response should contain "error"

  @failure @bulk @categorize @validation
  Scenario: Bulk categorize fails with non-existent category
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Test",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/transactions/bulk-categorize" with body:
      """
      {
        "ids": {{transaction_ids}},
        "category_id": "00000000-0000-0000-0000-000000000000"
      }
      """
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "TXN-010006"

  # ============================================
  # AUTHENTICATION SCENARIOS
  # ============================================

  @failure @unauthorized
  Scenario: Cannot access transactions without authentication
    Given the header is empty
    When I send a "GET" request to "/api/v1/transactions"
    Then the response status should be 401

  @failure @unauthorized
  Scenario: Cannot create transaction without authentication
    Given the header is empty
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "date": "2025-11-20",
        "description": "Test",
        "amount": -50.00,
        "type": "expense"
      }
      """
    Then the response status should be 401

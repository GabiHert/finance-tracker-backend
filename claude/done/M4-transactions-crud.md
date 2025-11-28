# Feature Planning Request

## Feature: Transaction Management CRUD API (M4-B1 to M4-B7)

## Description
Complete transaction management system with CRUD operations, filtering, pagination, and bulk operations for the Finance Tracker application.

## Specifications

### M4-B1: Transaction Domain Model (3 pts)
**Reference:** Backend TDD v6.0 Section 4.3

**Acceptance Criteria:**
- Transaction entity with: id, user_id, date, description, amount, type, category_id, notes, is_recurring, uploaded_at, timestamps
- Amount stored as decimal (15,2) - negative for expenses, positive for income
- Type enum: 'expense' | 'income'
- Soft-delete support with deleted_at
- Foreign key to categories table
- Foreign key to users table

### M4-B2: GET /transactions (5 pts)
**Reference:** Backend TDD v6.0 Section 4.3.1

**Query Parameters:**
- startDate, endDate (date range filter)
- categoryIds[] (multiple categories)
- type (expense | income)
- search (case-insensitive description match)
- page, limit (pagination, default 20 per page)
- groupByDate=true (groups by date with daily_total)

**Response:**
- transactions array with nested category
- pagination: page, limit, total, total_pages
- totals: income_total, expense_total, net_total

### M4-B3: POST /transactions (3 pts)
**Acceptance Criteria:**
- Create single transaction
- Validates required fields: date, description, amount, type
- Category optional (can be uncategorized)
- Returns created transaction with category populated
- Requires authentication

### M4-B4: PATCH /transactions/:id (3 pts)
**Acceptance Criteria:**
- Update transaction fields
- Validates ownership (user can only update own transactions)
- Returns updated transaction
- Supports partial updates

### M4-B5: DELETE /transactions/:id (2 pts)
**Acceptance Criteria:**
- Soft-delete transaction
- Validates ownership
- Returns 204 No Content

### M4-B6: POST /transactions/bulk-delete (3 pts)
**Acceptance Criteria:**
- Accept array of transaction IDs
- Validates ownership for all
- Atomic operation (all or nothing)
- Returns count of deleted transactions

### M4-B7: POST /transactions/bulk-categorize (3 pts)
**Acceptance Criteria:**
- Accept array of transaction IDs and category_id
- Validates ownership for all transactions
- Validates category ownership
- Atomic operation
- Returns count of updated transactions

## BDD Scenarios (Required First)

```gherkin
@all @transactions
Feature: Transaction Management
  As a user
  I want to manage my financial transactions
  So that I can track my income and expenses

  Background:
    Given the API server is running
    And a user exists with email "test@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    And a category exists with name "Food" and type "expense"

  @success @list
  Scenario: List transactions when none exist
    When I send a "GET" request to "/api/v1/transactions"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "transactions"
    And the response should contain "pagination"
    And the response should contain "totals"

  @success @list @filter
  Scenario: List transactions with date filter
    Given I have transactions for November 2025
    When I send a "GET" request to "/api/v1/transactions?startDate=2025-11-01&endDate=2025-11-30"
    Then the response status should be 200
    And all transactions should be within date range

  @success @list @filter
  Scenario: Filter transactions by type
    Given I have expense and income transactions
    When I send a "GET" request to "/api/v1/transactions?type=expense"
    Then the response status should be 200
    And all transactions should have type "expense"

  @success @list @search
  Scenario: Search transactions by description
    Given I have a transaction with description "UBER Trip"
    When I send a "GET" request to "/api/v1/transactions?search=uber"
    Then the response status should be 200
    And the transaction "UBER Trip" should be in results

  @success @create
  Scenario: Create expense transaction
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
    And the response field "description" should be "Grocery shopping"
    And the response field "amount" should be "-150.50"

  @success @create
  Scenario: Create income transaction
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
    And the response field "type" should be "income"

  @success @update
  Scenario: Update transaction description
    Given I have a transaction with description "Old description"
    When I send a "PATCH" request to "/api/v1/transactions/{{transaction_id}}" with body:
      """
      {
        "description": "New description"
      }
      """
    Then the response status should be 200
    And the response field "description" should be "New description"

  @success @delete
  Scenario: Delete transaction
    Given I have a transaction
    When I send a "DELETE" request to "/api/v1/transactions/{{transaction_id}}"
    Then the response status should be 204

  @failure @delete @not-found
  Scenario: Cannot delete non-existent transaction
    When I send a "DELETE" request to "/api/v1/transactions/00000000-0000-0000-0000-000000000000"
    Then the response status should be 404

  @success @bulk
  Scenario: Bulk delete transactions
    Given I have 3 transactions
    When I send a "POST" request to "/api/v1/transactions/bulk-delete" with body:
      """
      {
        "ids": ["{{tx1_id}}", "{{tx2_id}}", "{{tx3_id}}"]
      }
      """
    Then the response status should be 200
    And the response field "deleted_count" should be "3"

  @success @bulk
  Scenario: Bulk categorize transactions
    Given I have 2 uncategorized transactions
    When I send a "POST" request to "/api/v1/transactions/bulk-categorize" with body:
      """
      {
        "ids": ["{{tx1_id}}", "{{tx2_id}}"],
        "category_id": "{{category_id}}"
      }
      """
    Then the response status should be 200
    And the response field "updated_count" should be "2"

  @failure @validation
  Scenario: Validation error for missing date
    When I send a "POST" request to "/api/v1/transactions" with body:
      """
      {
        "description": "Test",
        "amount": -50,
        "type": "expense"
      }
      """
    Then the response status should be 400

  @failure @unauthorized
  Scenario: Cannot access transactions without authentication
    Given the header is empty
    When I send a "GET" request to "/api/v1/transactions"
    Then the response status should be 401
```

## Implementation Checklist
- [ ] Create BDD feature file first
- [ ] Implement Transaction entity in domain layer
- [ ] Create transaction errors in domain layer
- [ ] Create TransactionRepository interface in application layer
- [ ] Implement transaction use cases (List, Create, Update, Delete, BulkDelete, BulkCategorize)
- [ ] Create Transaction DTOs in integration layer
- [ ] Implement TransactionRepository in persistence layer
- [ ] Implement TransactionController
- [ ] Wire up routes and dependencies
- [ ] Run BDD tests until 100% pass

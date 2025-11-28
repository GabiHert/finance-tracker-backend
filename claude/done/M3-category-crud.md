# M3-B1 to M3-B5: Category CRUD Implementation

## Feature Overview
- **Task IDs:** M3-B1, M3-B2, M3-B3, M3-B4, M3-B5
- **Points:** 15 (3+3+3+3+3)
- **Domain:** Backend
- **Reference:** Backend TDD v6.0 Section 3.6 and Section 4.4

## Description
Complete Category CRUD implementation with domain model, repository, service, use cases, and REST endpoints. Categories can be owned by users or groups and support transaction statistics.

## Tasks Breakdown

### M3-B1: Category Domain Model (3 pts)
Create Category entity following Clean Architecture principles.

### M3-B2: GET /categories (3 pts)
List categories with optional filtering and transaction statistics.

### M3-B3: POST /categories (3 pts)
Create new category with validation.

### M3-B4: PATCH /categories/:id (3 pts)
Update existing category.

### M3-B5: DELETE /categories/:id (3 pts)
Delete category (soft delete or cascade handling).

## Category Entity Fields
| Field | Type | Validation | Default |
|-------|------|------------|---------|
| id | UUID | Required, auto-generated | - |
| name | string | Required, max 50 chars | - |
| color | string | Required, hex format (#XXXXXX) | #6366F1 |
| icon | string | Required, max 50 chars | tag |
| owner_type | OwnerType | Enum: user, group | - |
| owner_id | UUID | Required | - |
| type | CategoryType | Enum: expense, income | - |
| created_at | time.Time | Auto-generated | - |
| updated_at | time.Time | Auto-generated | - |

## API Endpoints

### GET /api/v1/categories
**Query Parameters:**
- `owner_type` (optional): "user" or "group"
- `owner_id` (optional): UUID of owner
- `type` (optional): "expense" or "income"
- `startDate` (optional): ISO date for statistics period
- `endDate` (optional): ISO date for statistics period

**Response (200):**
```json
{
  "categories": [
    {
      "id": "uuid",
      "name": "Food",
      "color": "#FF5733",
      "icon": "utensils",
      "owner_type": "user",
      "owner_id": "uuid",
      "type": "expense",
      "transaction_count": 15,
      "period_total": -1250.00,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

### POST /api/v1/categories
**Request:**
```json
{
  "name": "Food",
  "color": "#FF5733",
  "icon": "utensils",
  "owner_type": "user",
  "owner_id": "uuid",
  "type": "expense"
}
```

**Response (201):** Created category object

**Errors:**
- 400: Validation error
- 409: Category name already exists for this owner

### PATCH /api/v1/categories/:id
**Request:** Partial category object (name, color, icon)

**Response (200):** Updated category object

**Errors:**
- 400: Validation error
- 404: Category not found
- 403: Not owner of category

### DELETE /api/v1/categories/:id
**Response (204):** No content

**Errors:**
- 404: Category not found
- 403: Not owner of category

## Domain Errors
- CAT-010001: Category name too long
- CAT-010002: Invalid color format
- CAT-010003: Invalid owner type
- CAT-010004: Category not found
- CAT-010005: Category name already exists
- CAT-010006: Not authorized to modify category

## Implementation Requirements

### Domain Layer (`/internal/domain/`)
1. **Entity** (`/internal/domain/entity/category.go`):
   - Pure Go struct with all fields
   - No external dependencies
   - OwnerType and CategoryType enums

2. **Domain Errors** (`/internal/domain/error/category_errors.go`):
   - All CAT-01XXXX error codes

### Application Layer (`/internal/application/`)
1. **Adapter Interface** (`/internal/application/adapter/category_repository.go`):
   - FindByID, FindByOwner, Create, Update, Delete
   - FindByNameAndOwner (for uniqueness check)
   - GetTransactionStats (for computed fields)

2. **Use Cases** (`/internal/application/usecase/category/`):
   - list_categories.go
   - create_category.go
   - update_category.go
   - delete_category.go

### Integration Layer (`/internal/integration/`)
1. **Model** (`/internal/integration/persistence/model/category.go`):
   - GORM model with tags
   - ToEntity() and FromEntity() methods

2. **Repository** (`/internal/integration/persistence/category_repository.go`):
   - Implement adapter interface

3. **Controller** (`/internal/integration/entrypoint/controller/category.go`):
   - HTTP handlers for all endpoints

4. **DTOs** (`/internal/integration/entrypoint/dto/category.go`):
   - CreateCategoryRequest, UpdateCategoryRequest
   - CategoryResponse, CategoryListResponse

### Infrastructure Layer (`/internal/infra/`)
1. **Routes** (`/internal/infra/server/router/router.go`):
   - Add category routes under /api/v1/categories
   - All routes require authentication

## BDD Feature File Required
Create `/test/integration/features/categories.feature` with scenarios:

```gherkin
@all @categories
Feature: Category Management
  As a user
  I want to manage my transaction categories
  So that I can organize my financial data

  Background:
    Given I am authenticated as user "test@example.com"

  @success @list
  Scenario: List user categories
    Given I have the following categories:
      | name      | type    | color   | icon     |
      | Food      | expense | #FF5733 | utensils |
      | Transport | expense | #33FF57 | car      |
    When I GET "/api/v1/categories"
    Then the response status should be 200
    And the response should contain 2 categories

  @success @list @statistics
  Scenario: List categories with transaction statistics
    Given I have category "Food" with 5 transactions totaling -500.00
    When I GET "/api/v1/categories?startDate=2025-11-01&endDate=2025-11-30"
    Then the response status should be 200
    And category "Food" should have transaction_count 5
    And category "Food" should have period_total -500.00

  @success @create
  Scenario: Create new category
    When I POST "/api/v1/categories" with:
      | name  | Food     |
      | color | #FF5733  |
      | icon  | utensils |
      | type  | expense  |
    Then the response status should be 201
    And the response should contain "id"
    And the response should contain "name" "Food"

  @failure @create @duplicate
  Scenario: Cannot create duplicate category name
    Given I have category "Food"
    When I POST "/api/v1/categories" with:
      | name | Food    |
      | type | expense |
    Then the response status should be 409
    And the response should contain error "CAT-010005"

  @success @update
  Scenario: Update category
    Given I have category "Food" with id "cat-123"
    When I PATCH "/api/v1/categories/cat-123" with:
      | name  | Groceries |
      | color | #5733FF   |
    Then the response status should be 200
    And the response should contain "name" "Groceries"

  @failure @update @not-found
  Scenario: Cannot update non-existent category
    When I PATCH "/api/v1/categories/non-existent" with:
      | name | New Name |
    Then the response status should be 404

  @success @delete
  Scenario: Delete category
    Given I have category "Food" with id "cat-123"
    When I DELETE "/api/v1/categories/cat-123"
    Then the response status should be 204
    And category "cat-123" should not exist

  @failure @delete @not-found
  Scenario: Cannot delete non-existent category
    When I DELETE "/api/v1/categories/non-existent"
    Then the response status should be 404

  @failure @validation
  Scenario: Validation errors for category creation
    When I POST "/api/v1/categories" with:
      | name | |
    Then the response status should be 400
```

## Dependencies
- Existing auth middleware for authentication
- Database migrations 000004 (categories table)

## Files to Create/Modify
1. `/internal/domain/entity/category.go` - Category entity
2. `/internal/domain/error/category_errors.go` - Domain errors
3. `/internal/application/adapter/category_repository.go` - Repository interface
4. `/internal/application/usecase/category/` - All use cases
5. `/internal/integration/persistence/model/category.go` - GORM model
6. `/internal/integration/persistence/category_repository.go` - Repository impl
7. `/internal/integration/entrypoint/controller/category.go` - Controller
8. `/internal/integration/entrypoint/dto/category.go` - DTOs
9. `/internal/infra/server/router/router.go` - Add routes
10. `/internal/infra/dependency/injector.go` - Wire dependencies
11. `/test/integration/features/categories.feature` - BDD tests

## Notes
- Follow existing patterns in codebase (see auth implementation)
- Entity should be pure (no framework dependencies)
- All validation logic in use case layer
- Error codes follow PREFIX-XXYYYY format (CAT-01XXXX for category errors)
- Authentication required for all endpoints
- Owner verification required for update/delete operations

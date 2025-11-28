# M2-B1: User Domain Model

## Feature Overview
- **Task ID:** M2-B1
- **Points:** 3
- **Domain:** Backend
- **Reference:** Backend TDD v6.0 Section 3

## Description
Define User entity with all fields and validation rules following Clean Architecture principles.

## Acceptance Criteria
- User model includes all fields from User table schema
- Validation: email format, password requirements (min 8 chars), name (min 2 chars)
- Password hashing abstracted to service layer
- User preferences with sensible defaults (date_format, number_format, first_day_of_week)
- No password included in public user responses

## User Entity Fields
| Field | Type | Validation | Default |
|-------|------|------------|---------|
| id | UUID | Required, auto-generated | - |
| email | string | Required, valid email format, unique | - |
| name | string | Required, min 2 chars | - |
| password_hash | string | Required (internal only) | - |
| date_format | DateFormat | Enum: YYYY-MM-DD, DD/MM/YYYY, MM/DD/YYYY | YYYY-MM-DD |
| number_format | NumberFormat | Enum: US, EU, BR | US |
| first_day_of_week | FirstDayOfWeek | Enum: sunday, monday | sunday |
| email_notifications | bool | - | true |
| goal_alerts | bool | - | true |
| recurring_reminders | bool | - | true |
| terms_accepted_at | time.Time | Required | - |
| created_at | time.Time | Required | - |
| updated_at | time.Time | Required | - |

## Implementation Requirements

### Domain Layer (`/internal/domain/`)
1. **Entity** (`/internal/domain/entity/user.go`):
   - Pure Go struct with all fields
   - No external dependencies
   - Type definitions for enums (DateFormat, NumberFormat, FirstDayOfWeek)

2. **Domain Errors** (`/internal/domain/error/`):
   - AUTH-010001: Invalid email format
   - AUTH-010002: Password too short
   - AUTH-010003: Name too short
   - AUTH-010004: Terms not accepted
   - AUTH-010005: Email already exists

3. **Validation Rules**:
   - Email: Valid format using regex
   - Password: Minimum 8 characters
   - Name: Minimum 2 characters
   - Terms: Must be accepted (terms_accepted_at not zero)

### Application Layer (`/internal/application/`)
1. **Adapter Interface** (`/internal/application/adapter/user_repository.go`):
   - Already exists, verify it has all needed methods
   - FindByID, FindByEmail, Create, Update, Delete

2. **User Service Interface** (if not exists):
   - CreateUser, GetUser, UpdateUser, etc.

### Integration Layer (`/internal/integration/`)
1. **Model** (`/internal/integration/persistence/model/user.go`):
   - Already exists with GORM tags
   - Has ToEntity() and FromEntity() methods

## BDD Feature File Required
Create `/test/integration/features/user-domain.feature` with scenarios:
- User entity creation with valid data
- User entity validation failures
- Email format validation
- Password length validation
- Name length validation

## Dependencies
- github.com/google/uuid (already in go.mod)
- No external validation library (use stdlib)

## Files to Create/Modify
1. `/internal/domain/entity/user.go` - Verify/enhance User entity
2. `/internal/domain/error/auth_errors.go` - Create auth-specific errors
3. `/internal/application/adapter/user_repository.go` - Verify interface
4. `/test/integration/features/user-domain.feature` - BDD tests

## Notes
- Follow existing patterns in codebase
- Entity should be pure (no framework dependencies)
- All validation logic in service layer, not entity
- Error codes follow PREFIX-XXYYYY format (AUTH-01XXXX for auth errors)

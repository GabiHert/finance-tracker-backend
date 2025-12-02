# Feature Planning Request

## Feature: Goals (Spending Limits) API (M7-B1 to M7-B6)

## Description
Complete goal/spending limit management system with CRUD operations for the Finance Tracker application. Goals allow users to set monthly spending limits per category and track their progress against those limits.

## Specifications

### M7-B1: Goal Domain Model
**Acceptance Criteria:**
- Goal entity with: id, user_id, category_id, limit_amount, alert_on_exceed, period, start_date, end_date, timestamps
- limit_amount stored as decimal (15,2) - positive values only
- period enum: 'monthly' | 'weekly' | 'yearly' (default: 'monthly')
- Foreign key to categories table
- Foreign key to users table
- Unique constraint on (user_id, category_id) - one goal per category per user

### M7-B2: GET /api/v1/goals
List all goals for the authenticated user.
- Response: Array of goals with category details and current spending
- Each goal includes: id, category_id, category (nested), limit_amount, current_amount, alert_on_exceed, period, start_date, end_date, created_at, updated_at
- current_amount is calculated from transactions within the period

### M7-B3: POST /api/v1/goals
Create a new spending goal.
- Request: { category_id, limit_amount, alert_on_exceed?, period? }
- Validates category exists and belongs to user
- Validates limit_amount > 0
- Returns created goal with category details

### M7-B4: GET /api/v1/goals/:id
Get a single goal by ID.
- Response: Goal with category details and current spending
- Validates goal belongs to authenticated user

### M7-B5: PATCH /api/v1/goals/:id
Update an existing goal.
- Request: { limit_amount?, alert_on_exceed?, period? }
- Note: category_id cannot be changed (delete and recreate instead)
- Validates goal belongs to authenticated user
- Returns updated goal

### M7-B6: DELETE /api/v1/goals/:id
Delete a goal.
- Validates goal belongs to authenticated user
- Returns 204 No Content

## Database Schema

### goals table
```sql
CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    limit_amount DECIMAL(15,2) NOT NULL,
    alert_on_exceed BOOLEAN NOT NULL DEFAULT true,
    period VARCHAR(20) NOT NULL DEFAULT 'monthly' CHECK (period IN ('monthly', 'weekly', 'yearly')),
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(user_id, category_id, deleted_at)
);

CREATE INDEX idx_goals_user_id ON goals(user_id);
CREATE INDEX idx_goals_category_id ON goals(category_id);
```

## BDD Scenarios

To be created in `/test/integration/features/goals.feature`

## Implementation Checklist
- [ ] Create BDD feature file first
- [ ] Create database migration for goals table
- [ ] Implement Goal entity in domain layer
- [ ] Create goal errors in domain layer (GOL-XXXXXX format)
- [ ] Create GoalRepository interface in application layer
- [ ] Implement goal use cases (List, Create, Get, Update, Delete)
- [ ] Create Goal DTOs in integration layer
- [ ] Implement GoalRepository in persistence layer
- [ ] Implement GoalController
- [ ] Wire up routes and dependencies
- [ ] Run BDD tests until 100% pass

## Error Codes
Following the project convention (PREFIX-XXYYYY):
- GOL-010001: Goal not found
- GOL-010002: Goal already exists for this category
- GOL-010003: Invalid limit amount (must be > 0)
- GOL-010004: Category not found
- GOL-010005: Category does not belong to user
- GOL-010006: Unauthorized access to goal

## Existing Code to Leverage
- Transaction repository pattern for current_amount calculation
- Category repository for validation
- Auth middleware for user authentication
- Existing BDD step definitions (see step_definitions_test.go)

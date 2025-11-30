# M9: Groups & Collaboration Feature

## Feature Overview
Implementation of Groups & Collaboration functionality for the Finance Tracker application.

## Requirements (from Finance-Tracker-Implementation-Guide-Complete-v1.md)

### Database Schema (M9-I1, M9-I2)

#### groups table
- id: UUID PRIMARY KEY
- name: VARCHAR(100) NOT NULL
- created_by: UUID NOT NULL REFERENCES users(id)
- created_at: TIMESTAMP NOT NULL DEFAULT NOW()
- updated_at: TIMESTAMP NOT NULL DEFAULT NOW()

#### group_members table
- id: UUID PRIMARY KEY
- group_id: UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE
- user_id: UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
- role: VARCHAR(20) NOT NULL (admin | member)
- joined_at: TIMESTAMP NOT NULL DEFAULT NOW()
- UNIQUE(group_id, user_id)

#### group_invites table
- id: UUID PRIMARY KEY
- group_id: UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE
- email: VARCHAR(255) NOT NULL
- token: VARCHAR(64) NOT NULL UNIQUE
- invited_by: UUID NOT NULL REFERENCES users(id)
- status: VARCHAR(20) NOT NULL DEFAULT 'pending' (pending | accepted | declined | expired)
- expires_at: TIMESTAMP NOT NULL
- created_at: TIMESTAMP NOT NULL DEFAULT NOW()

### API Endpoints

#### M9-B2: POST /api/v1/groups
Create a new group. User becomes admin automatically.
- Request: { name: string }
- Response: { id, name, created_at, members: [{ id, name, email, role }] }

#### M9-B3: GET /api/v1/groups
List all groups user belongs to.
- Response: [{ id, name, member_count, role, created_at }]

#### M9-B4: GET /api/v1/groups/:id
Get group details including members.
- Response: { id, name, members: [...], pending_invites: [...], created_at }

#### M9-B5: POST /api/v1/groups/:id/invite
Send invitation to join group (admin only).
- Request: { email: string }
- Response: { invite_id, email, status, expires_at }

#### M9-B6: POST /api/v1/groups/invites/:token/accept
Accept an invitation using token.
- Response: { group_id, group_name }

#### M9-B7: Member Management Endpoints
- PUT /api/v1/groups/:id/members/:member_id/role - Change member role (admin only)
  - Request: { role: 'admin' | 'member' }
- DELETE /api/v1/groups/:id/members/:member_id - Remove member (admin only)
- DELETE /api/v1/groups/:id/members/me - Leave group (self)

#### M9-B8: GET /api/v1/groups/:id/transactions
Get group transactions (all members' transactions tagged for this group).

#### M9-B9: GET /api/v1/groups/:id/dashboard
Get group dashboard metrics (totals, spending by member, spending by category).

## BDD Scenarios

The following scenarios should be implemented in `/test/integration/features/groups.feature`:

```gherkin
Feature: Groups & Collaboration
  As a user
  I want to create and manage groups
  So that I can collaborate on finances with family/friends

  Background:
    Given the database is clean
    And I have a valid authentication token

  Scenario: Create a new group
    When I send a POST request to "/api/v1/groups" with:
      | name | Familia Silva |
    Then the response status should be 201
    And the response should contain "Familia Silva"
    And a group should exist with name "Familia Silva"
    And I should be a member of the group with role "admin"

  Scenario: List my groups
    Given I am a member of group "Familia" with role "admin"
    And I am a member of group "Amigos" with role "member"
    When I send a GET request to "/api/v1/groups"
    Then the response status should be 200
    And the response should contain 2 groups

  Scenario: Get group details
    Given I am a member of group "Familia" with role "admin"
    And the group has member "maria@example.com" with role "member"
    When I send a GET request to "/api/v1/groups/{group_id}"
    Then the response status should be 200
    And the response should contain 2 members

  Scenario: Invite member to group as admin
    Given I am a member of group "Familia" with role "admin"
    When I send a POST request to "/api/v1/groups/{group_id}/invite" with:
      | email | maria@example.com |
    Then the response status should be 201
    And a pending invite should exist for "maria@example.com"

  Scenario: Cannot invite member as non-admin
    Given I am a member of group "Familia" with role "member"
    When I send a POST request to "/api/v1/groups/{group_id}/invite" with:
      | email | pedro@example.com |
    Then the response status should be 403

  Scenario: Accept group invitation
    Given I have a pending invite to group "Familia" with token "valid-token"
    When I send a POST request to "/api/v1/groups/invites/valid-token/accept"
    Then the response status should be 200
    And I should be a member of the group "Familia"

  Scenario: Change member role as admin
    Given I am a member of group "Familia" with role "admin"
    And the group has member "maria@example.com" with role "member"
    When I send a PUT request to "/api/v1/groups/{group_id}/members/{member_id}/role" with:
      | role | admin |
    Then the response status should be 200
    And "maria@example.com" should have role "admin"

  Scenario: Remove member as admin
    Given I am a member of group "Familia" with role "admin"
    And the group has member "pedro@example.com" with role "member"
    When I send a DELETE request to "/api/v1/groups/{group_id}/members/{member_id}"
    Then the response status should be 204
    And "pedro@example.com" should not be a member of the group

  Scenario: Leave group as member
    Given I am a member of group "Amigos" with role "member"
    When I send a DELETE request to "/api/v1/groups/{group_id}/members/me"
    Then the response status should be 204
    And I should not be a member of the group

  Scenario: Admin cannot leave if sole admin
    Given I am the only admin of group "Familia"
    And the group has other members
    When I send a DELETE request to "/api/v1/groups/{group_id}/members/me"
    Then the response status should be 400
    And the response should contain error "Cannot leave: you are the only admin"
```

## Implementation Order

1. Database migrations (M9-I1, M9-I2)
2. Domain entities (Group, GroupMember, GroupInvite)
3. Repository interfaces and implementations
4. Use cases (CreateGroup, ListGroups, GetGroup, InviteMember, AcceptInvite, ChangeMemberRole, RemoveMember, LeaveGroup)
5. HTTP handlers/controllers
6. Wire routes in router

## Acceptance Criteria

- [ ] Groups can be created with the creator as admin
- [ ] Users can list all groups they belong to
- [ ] Group details show all members and pending invites
- [ ] Admins can invite new members via email
- [ ] Invited users can accept/decline invitations
- [ ] Admins can change member roles
- [ ] Admins can remove members
- [ ] Members can leave groups (except sole admin)
- [ ] All endpoints require authentication
- [ ] All BDD scenarios pass

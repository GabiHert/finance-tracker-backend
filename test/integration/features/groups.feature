@all @groups
Feature: Groups & Collaboration
  As a user
  I want to create and manage groups
  So that I can collaborate on finances with family/friends

  Background:
    Given the API server is running
    And a user exists with email "test@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens

  # ==================== Group Creation ====================

  @success @create
  Scenario: Create a new group
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia Silva"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "name" should be "Familia Silva"
    And the response field "id" should exist
    And the response should contain "members"

  @success @create
  Scenario: Create group and verify creator is admin
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "My Group"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "members.0.role" should be "admin"
    And the response field "members.0.email" should be "test@example.com"

  @failure @create @validation
  Scenario: Cannot create group without name
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {}
      """
    Then the response status should be 400
    And the response should be JSON

  @failure @create @validation
  Scenario: Cannot create group with name too long
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "This is a very long group name that exceeds one hundred characters and should not be allowed by the system validation"
      }
      """
    Then the response status should be 400
    And the response should be JSON

  # ==================== List Groups ====================

  @success @list
  Scenario: List user groups when none exist
    When I send a "GET" request to "/api/v1/groups"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "groups"

  @success @list
  Scenario: List user groups after creating one
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "GET" request to "/api/v1/groups"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "groups"

  # ==================== Get Group Details ====================

  @success @get
  Scenario: Get group details
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    And the response field "id" should exist
    When I send a "GET" request to "/api/v1/groups/{{group_id}}"
    Then the response status should be 200
    And the response should be JSON
    And the response field "name" should be "Familia"
    And the response should contain "members"

  @failure @get @not-found
  Scenario: Cannot get non-existent group
    When I send a "GET" request to "/api/v1/groups/00000000-0000-0000-0000-000000000000"
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "GRP-010001"

  # ==================== Invite Member ====================

  @success @invite
  Scenario: Invite member to group as admin
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "maria@example.com"
      }
      """
    Then the response status should be 201
    And the response should be JSON
    And the response field "email" should be "maria@example.com"
    And the response field "status" should be "pending"
    And the response field "token" should exist

  @failure @invite @validation
  Scenario: Cannot invite with invalid email
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "invalid-email"
      }
      """
    Then the response status should be 400
    And the response should be JSON

  @failure @invite @duplicate
  Scenario: Cannot invite already invited user
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "maria@example.com"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "maria@example.com"
      }
      """
    Then the response status should be 409
    And the response should be JSON
    And the response field "code" should be "GRP-010005"

  # ==================== Accept Invite ====================

  @success @accept-invite
  Scenario: Accept group invitation
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "maria@example.com"
      }
      """
    Then the response status should be 201
    And the response field "token" should exist
    When I send a "POST" request to "/api/v1/groups/invites/{{invite_token}}/accept"
    Then the response status should be 200
    And the response should be JSON
    And the response field "group_id" should exist
    And the response field "group_name" should be "Familia"

  @failure @accept-invite @invalid-token
  Scenario: Cannot accept with invalid token
    When I send a "POST" request to "/api/v1/groups/invites/invalid-token-here/accept"
    Then the response status should be 404
    And the response should be JSON
    And the response field "code" should be "GRP-010006"

  # ==================== Change Member Role ====================

  @success @change-role
  Scenario: Change member role as admin
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "maria@example.com"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/invites/{{invite_token}}/accept"
    Then the response status should be 200
    When I send a "PUT" request to "/api/v1/groups/{{group_id}}/members/{{member_id}}/role" with body:
      """
      {
        "role": "admin"
      }
      """
    Then the response status should be 200
    And the response should be JSON
    And the response field "role" should be "admin"

  @failure @change-role @validation
  Scenario: Cannot set invalid role
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "maria@example.com"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/invites/{{invite_token}}/accept"
    Then the response status should be 200
    When I send a "PUT" request to "/api/v1/groups/{{group_id}}/members/{{member_id}}/role" with body:
      """
      {
        "role": "superadmin"
      }
      """
    Then the response status should be 400
    And the response should be JSON

  # ==================== Remove Member ====================

  @success @remove-member
  Scenario: Remove member as admin
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "maria@example.com"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/invites/{{invite_token}}/accept"
    Then the response status should be 200
    When I send a "DELETE" request to "/api/v1/groups/{{group_id}}/members/{{member_id}}"
    Then the response status should be 204

  @failure @remove-member @not-found
  Scenario: Cannot remove non-existent member
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "DELETE" request to "/api/v1/groups/{{group_id}}/members/00000000-0000-0000-0000-000000000000"
    Then the response status should be 404
    And the response should be JSON

  # ==================== Leave Group ====================

  @success @leave
  Scenario: Leave group as member
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "maria@example.com"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/invites/{{invite_token}}/accept"
    Then the response status should be 200
    When I send a "PUT" request to "/api/v1/groups/{{group_id}}/members/{{member_id}}/role" with body:
      """
      {
        "role": "admin"
      }
      """
    Then the response status should be 200
    When I send a "DELETE" request to "/api/v1/groups/{{group_id}}/members/me"
    Then the response status should be 204

  @failure @leave @sole-admin
  Scenario: Cannot leave as sole admin
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Familia"
      }
      """
    Then the response status should be 201
    When I send a "DELETE" request to "/api/v1/groups/{{group_id}}/members/me"
    Then the response status should be 400
    And the response should be JSON
    And the response field "code" should be "GRP-010008"

  # ==================== Authentication ====================

  @failure @unauthorized
  Scenario: Cannot access groups without authentication
    Given the header is empty
    When I send a "GET" request to "/api/v1/groups"
    Then the response status should be 401

  @failure @unauthorized
  Scenario: Cannot create group without authentication
    Given the header is empty
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Test Group"
      }
      """
    Then the response status should be 401

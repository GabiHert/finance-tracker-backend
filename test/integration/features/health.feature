# Finance Tracker - Health Check Feature
# Version: 1.0 | Milestone 1 (M1-B1, M1-B2)

Feature: API Health Check
  As a system operator
  I want to check if the API is healthy
  So that I can ensure the service is running properly

  Background:
    Given the API server is running

  @all @health @success
  Scenario: Health endpoint returns OK when server is running
    When I send a "GET" request to "/health"
    Then the response status should be 200
    And the response should be JSON
    And the response should contain "status"
    And the response field "status" should be "ok"

  @all @health @success
  Scenario: Health endpoint returns database status
    When I send a "GET" request to "/health"
    Then the response status should be 200
    And the response should contain "database"

  @all @health @success
  Scenario: Health endpoint returns timestamp
    When I send a "GET" request to "/health"
    Then the response status should be 200
    And the response should contain "timestamp"

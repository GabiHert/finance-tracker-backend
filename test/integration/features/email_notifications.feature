# Finance Tracker - Email Notifications Feature
# Version: 1.0 | Milestone 14 (M14-email-notifications)

@all @email
Feature: Email Notifications
  As a system
  I want to queue and send email notifications
  So that users receive important communications reliably

  Background:
    Given the API server is running

  # ============================================
  # PASSWORD RESET EMAIL SCENARIOS
  # ============================================

  @email @password-reset @success
  Scenario: Password reset request queues an email
    Given a user exists with email "resetuser@example.com" and password "SecurePass123!"
    When I send a "POST" request to "/api/v1/auth/forgot-password" with body:
      """
      {
        "email": "resetuser@example.com"
      }
      """
    Then the response status should be 200
    And an email job should be queued for "resetuser@example.com"
    And the email job should have template type "password_reset"
    And the email job should have status "pending"

  @email @password-reset @success
  Scenario: Password reset email contains correct template data
    Given a user exists with email "templatetest@example.com" and password "SecurePass123!"
    When I send a "POST" request to "/api/v1/auth/forgot-password" with body:
      """
      {
        "email": "templatetest@example.com"
      }
      """
    Then the response status should be 200
    And an email job should be queued for "templatetest@example.com"
    And the email job template data should contain "user_name"
    And the email job template data should contain "reset_url"
    And the email job template data should contain "expires_in"

  @email @password-reset @no-email
  Scenario: Password reset for non-existent email does not queue email
    When I send a "POST" request to "/api/v1/auth/forgot-password" with body:
      """
      {
        "email": "nonexistent@example.com"
      }
      """
    Then the response status should be 200
    And no email job should be queued for "nonexistent@example.com"

  # ============================================
  # GROUP INVITATION EMAIL SCENARIOS
  # ============================================

  @email @group-invitation @success
  Scenario: Group invitation queues an email
    Given a user exists with email "inviter@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Test Family Group"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "newmember@example.com",
        "confirm_non_user": true
      }
      """
    Then the response status should be 201
    And an email job should be queued for "newmember@example.com"
    And the email job should have template type "group_invitation"
    And the email job should have status "pending"

  @email @group-invitation @success
  Scenario: Group invitation email contains correct template data
    Given a user exists with email "inviter2@example.com" and password "SecurePass123!"
    And the user is logged in with valid tokens
    When I send a "POST" request to "/api/v1/groups" with body:
      """
      {
        "name": "Family Budget"
      }
      """
    Then the response status should be 201
    When I send a "POST" request to "/api/v1/groups/{{group_id}}/invite" with body:
      """
      {
        "email": "invitee@example.com",
        "confirm_non_user": true
      }
      """
    Then the response status should be 201
    And an email job should be queued for "invitee@example.com"
    And the email job template data should contain "inviter_name"
    And the email job template data should contain "inviter_email"
    And the email job template data should contain "group_name"
    And the email job template data should contain "invite_url"
    And the email job template data should contain "expires_in"

  # ============================================
  # EMAIL QUEUE PROCESSING SCENARIOS
  # ============================================

  @email @worker @success
  Scenario: Email worker processes pending emails
    Given a pending email job exists for "worker-test@example.com"
    When the email worker processes the queue
    Then the email job for "worker-test@example.com" should have status "sent"
    And the email job should have a resend_id

  @email @worker @retry
  Scenario: Email worker retries on temporary failure
    Given a pending email job exists for "retry-test@example.com"
    And the email sender will fail with a temporary error
    When the email worker processes the queue
    Then the email job for "retry-test@example.com" should have status "pending"
    And the email job should have attempts equal to 1
    And the email job should have scheduled_at in the future

  @email @worker @failure
  Scenario: Email marked as failed after max retries
    Given an email job with 2 failed attempts exists for "failed-test@example.com"
    And the email sender will fail with a temporary error
    When the email worker processes the queue
    Then the email job for "failed-test@example.com" should have status "failed"
    And the email job should have attempts equal to 3
    And the email job should have a last_error

  @email @worker @permanent-failure
  Scenario: Permanent errors are not retried
    Given a pending email job exists for "permanent-fail@example.com"
    And the email sender will fail with a permanent error
    When the email worker processes the queue
    Then the email job for "permanent-fail@example.com" should have status "failed"
    And the email job should have attempts equal to 1

  # ============================================
  # EMAIL QUEUE DATABASE SCENARIOS
  # ============================================

  @email @database @success
  Scenario: Email job is persisted with correct fields
    Given a user exists with email "persist-test@example.com" and password "SecurePass123!"
    When I send a "POST" request to "/api/v1/auth/forgot-password" with body:
      """
      {
        "email": "persist-test@example.com"
      }
      """
    Then the response status should be 200
    And the database should contain an email_queue record with:
      | recipient_email | persist-test@example.com |
      | template_type   | password_reset           |
      | status          | pending                  |
      | attempts        | 0                        |
      | max_attempts    | 3                        |

# Feature: M14-email-notifications

## Description
The email notification system provides reliable, asynchronous email delivery for critical user flows including password reset and group invitation emails. It uses a database-backed queue with retry logic and integrates with Resend for email delivery.

## User Story
As a system administrator, I want users to receive professional email notifications for critical actions so that they can complete password resets and respond to group invitations.

## Business Value
- Enables password recovery flow for users who forget their credentials
- Facilitates collaboration through group invitation notifications
- Provides reliable email delivery with automatic retry on failures
- Supports future email notification types (goal alerts, weekly reports)

---

## Functional Requirements

### Core Functionality
- Database-backed email queue with status tracking (pending, processing, sent, failed)
- Background worker that polls queue every 5 seconds and processes up to 10 emails per batch
- Automatic retry with exponential backoff (0s, 1min, 5min) up to 3 attempts
- HTML and plain text email template rendering with Go templates
- Integration with Resend API for email delivery

### Email Types (Phase 1)
1. **Password Reset Email**
   - Template: `password_reset`
   - Subject: "Redefinir sua senha - Finance Tracker"
   - Data: user_name, reset_url, expires_in

2. **Group Invitation Email**
   - Template: `group_invitation`
   - Subject: "{InviterName} convidou voce para {GroupName} - Finance Tracker"
   - Data: inviter_name, inviter_email, group_name, invite_url, expires_in

### Data Requirements
- EmailJob entity with: id, template_type, recipient_email, recipient_name, subject, template_data (JSONB), status, attempts, max_attempts, last_error, resend_id, created_at, scheduled_at, processed_at
- Email queue repository for CRUD operations
- Template renderer for HTML/text email generation

---

## Definition of Done

### Functionality Checklist
- [ ] Database migration creates email_queue table with indexes
- [ ] EmailJob domain entity with status management methods
- [ ] EmailQueueRepository interface and implementation
- [ ] Email service for queueing emails
- [ ] Template renderer for password_reset and group_invitation
- [ ] Background worker with polling and batch processing
- [ ] Resend client integration
- [ ] ForgotPassword usecase queues password reset email
- [ ] InviteMember usecase queues group invitation email
- [ ] Retry logic with exponential backoff
- [ ] Proper error handling for permanent vs temporary failures

### Code Quality Checklist
- [ ] Code follows all rules in CLAUDE.md
- [ ] Clean architecture layers respected (domain -> application -> integration -> infra)
- [ ] No business logic in controllers or repositories
- [ ] Proper dependency injection via constructors
- [ ] Error wrapping with context

### Testing Checklist
- [ ] BDD feature file for email queue scenarios
- [ ] BDD feature file for email worker scenarios
- [ ] BDD feature file for password reset email flow
- [ ] BDD feature file for group invitation email flow
- [ ] Mock email sender for testing

### Documentation Checklist
- [ ] GoDoc comments for all exported types and functions
- [ ] Template documentation

---

## Security Requirements

### Data Protection
- Email addresses validated before queueing
- Template data sanitized before rendering
- Resend API key stored in environment variables
- No sensitive data (passwords, tokens) logged

### Error Handling
- Permanent errors (401, 422) not retried
- Temporary errors (429, 5xx) retried with backoff
- Error messages sanitized before storage

---

## Performance Specifications

### Worker Performance
- Poll interval: 5 seconds (configurable)
- Batch size: 10 emails per cycle (configurable)
- Non-blocking queue operations
- Graceful shutdown support

### Database Performance
- Index on (status, scheduled_at) for pending emails
- Index on recipient_email for lookups
- JSONB for flexible template data

---

## Implementation Plan

### Phase 1: Database and Domain
1. Create database migration for email_queue table
2. Implement EmailJob domain entity with status methods
3. Implement domain errors for email operations

### Phase 2: Repository Layer
4. Define EmailQueueRepository adapter interface
5. Implement PostgreSQL email queue repository

### Phase 3: Email Infrastructure
6. Implement Resend client adapter
7. Create HTML/text email templates
8. Implement template renderer

### Phase 4: Email Service and Worker
9. Implement email service for queueing
10. Implement background worker with polling

### Phase 5: Integration with Existing Usecases
11. Add EmailService dependency to ForgotPassword usecase
12. Add EmailService dependency to InviteMember usecase
13. Wire dependencies in injector

### Phase 6: Testing
14. Create BDD feature files
15. Implement mock email sender
16. Run all BDD tests

---

## Dependencies and Reusable Components

### New Third-party Dependencies
- `github.com/resend/resend-go/v2` - Resend SDK

### Existing Components to Integrate With
- `internal/application/usecase/auth/forgot_password.go` - Password reset flow
- `internal/application/usecase/group/invite_member.go` - Group invitation flow
- `internal/infra/dependency/injector.go` - Dependency injection

### Environment Variables
```
RESEND_API_KEY=re_xxxxxxxxxxxxx
RESEND_FROM_NAME=Finance Tracker
RESEND_FROM_EMAIL=onboarding@resend.dev
APP_BASE_URL=http://localhost:5173
EMAIL_WORKER_POLL_INTERVAL=5s
EMAIL_WORKER_BATCH_SIZE=10
EMAIL_WORKER_ENABLED=true
```

---

## Files to Create

```
backend/
├── internal/
│   ├── domain/
│   │   ├── entity/email_job.go
│   │   └── errors/email_errors.go
│   ├── application/adapter/
│   │   ├── email_queue_repository.go
│   │   └── email_sender.go
│   └── integration/
│       ├── email/
│       │   ├── service.go
│       │   ├── worker.go
│       │   ├── resend_client.go
│       │   └── templates/
│       │       ├── renderer.go
│       │       ├── password_reset.html
│       │       ├── password_reset.txt
│       │       ├── group_invitation.html
│       │       └── group_invitation.txt
│       └── persistence/email_queue_repository.go
├── scripts/migrations/
│   ├── 000XXX_create_email_queue.up.sql
│   └── 000XXX_create_email_queue.down.sql
└── test/
    ├── integration/features/email_notifications.feature
    └── mocks/email_sender.go
```

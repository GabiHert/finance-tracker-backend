-- Migration: Create email_queue table
-- Feature: M14-email-notifications

CREATE TABLE IF NOT EXISTS email_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Email content
    template_type VARCHAR(50) NOT NULL,
    recipient_email VARCHAR(255) NOT NULL,
    recipient_name VARCHAR(255),
    subject VARCHAR(500) NOT NULL,
    template_data JSONB NOT NULL DEFAULT '{}',

    -- Processing state
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    last_error TEXT,

    -- Resend tracking
    resend_id VARCHAR(100),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT email_queue_valid_status CHECK (status IN ('pending', 'processing', 'sent', 'failed')),
    CONSTRAINT email_queue_valid_template CHECK (template_type IN ('password_reset', 'group_invitation')),
    CONSTRAINT email_queue_valid_email CHECK (recipient_email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Index for worker to find pending jobs efficiently
CREATE INDEX idx_email_queue_pending ON email_queue(scheduled_at)
    WHERE status = 'pending';

-- Index for looking up emails by recipient (debugging, testing)
CREATE INDEX idx_email_queue_recipient ON email_queue(recipient_email, created_at DESC);

-- Index for monitoring failed emails
CREATE INDEX idx_email_queue_failed ON email_queue(created_at DESC)
    WHERE status = 'failed';

-- Comments
COMMENT ON TABLE email_queue IS 'Queue for outgoing email notifications';
COMMENT ON COLUMN email_queue.template_type IS 'Email template identifier (password_reset, group_invitation)';
COMMENT ON COLUMN email_queue.template_data IS 'JSON data for template rendering';
COMMENT ON COLUMN email_queue.status IS 'pending=waiting, processing=sending, sent=delivered, failed=gave up';
COMMENT ON COLUMN email_queue.resend_id IS 'Resend API email ID for tracking';

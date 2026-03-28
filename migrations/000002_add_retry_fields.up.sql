ALTER TYPE notification_status ADD VALUE IF NOT EXISTS 'dead_letter';

ALTER TABLE notifications
    ADD COLUMN retry_count   INT         NOT NULL DEFAULT 0,
    ADD COLUMN max_retries   INT         NOT NULL DEFAULT 5,
    ADD COLUMN next_retry_at TIMESTAMPTZ;

CREATE INDEX idx_notifications_retry ON notifications (next_retry_at)
    WHERE status = 'failed' AND next_retry_at IS NOT NULL;

ALTER TABLE notifications
    DROP COLUMN IF EXISTS retry_count,
    DROP COLUMN IF EXISTS max_retries,
    DROP COLUMN IF EXISTS next_retry_at;

DROP INDEX IF EXISTS idx_notifications_retry;

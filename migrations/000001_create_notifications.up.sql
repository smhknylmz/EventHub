CREATE TYPE notification_status AS ENUM (
    'pending', 'queued', 'processing',
    'delivered', 'failed', 'cancelled', 'dead_letter'
);

CREATE TYPE notification_channel AS ENUM ('email', 'sms', 'push');

CREATE TYPE notification_priority AS ENUM ('high', 'normal', 'low');

CREATE TABLE notifications (
    id              UUID PRIMARY KEY,
    batch_id        UUID,
    recipient       VARCHAR(255)         NOT NULL,
    channel         notification_channel NOT NULL,
    content         TEXT                 NOT NULL,
    priority        notification_priority NOT NULL DEFAULT 'normal',
    status          notification_status  NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ          NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_status ON notifications (status);
CREATE INDEX idx_notifications_batch_id ON notifications (batch_id) WHERE batch_id IS NOT NULL;
CREATE INDEX idx_notifications_channel_status ON notifications (channel, status);
CREATE INDEX idx_notifications_created_at ON notifications (created_at);

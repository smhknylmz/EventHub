DROP INDEX IF EXISTS idx_notifications_template_id;
ALTER TABLE notifications DROP COLUMN IF EXISTS template_id;

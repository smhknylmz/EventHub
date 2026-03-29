ALTER TABLE notifications ADD COLUMN template_id UUID REFERENCES templates(id) ON DELETE SET NULL;
CREATE INDEX idx_notifications_template_id ON notifications (template_id) WHERE template_id IS NOT NULL;

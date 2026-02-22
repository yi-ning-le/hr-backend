CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    subject_type VARCHAR(64) NOT NULL,
    subject_id UUID NOT NULL,
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_notifications_event_type
        CHECK (event_type IN ('candidate_reviewer_assigned', 'interview_assigned')),
    CONSTRAINT chk_notifications_subject_type
        CHECK (subject_type IN ('candidate', 'interview'))
);

CREATE INDEX idx_notifications_user_created_at
ON notifications (user_id, created_at DESC);

CREATE INDEX idx_notifications_user_unread
ON notifications (user_id, read_at)
WHERE read_at IS NULL;

CREATE INDEX idx_notifications_subject
ON notifications (subject_type, subject_id);


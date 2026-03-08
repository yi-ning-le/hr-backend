ALTER TABLE notifications
DROP CONSTRAINT IF EXISTS chk_notifications_event_type;

ALTER TABLE notifications
ADD CONSTRAINT chk_notifications_event_type
CHECK (event_type IN ('candidate_reviewer_assigned', 'interview_assigned', 'review_completed', 'interview_completed'));

-- Add interview_pass and interview_fail to allowed comment types
ALTER TABLE candidate_comments
DROP CONSTRAINT IF EXISTS chk_candidate_comments_type;

ALTER TABLE candidate_comments
ADD CONSTRAINT chk_candidate_comments_type
CHECK (comment_type IN ('normal', 'review_suitable', 'review_unsuitable', 'interview_pass', 'interview_fail'));

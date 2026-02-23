-- Add assigned_by_user_id to candidate_reviewers for tracking who assigned the reviewer
ALTER TABLE candidate_reviewers
ADD COLUMN IF NOT EXISTS assigned_by_user_id UUID REFERENCES users(id);

-- Add comment_type to candidate_comments for review decision comments
ALTER TABLE candidate_comments
ADD COLUMN IF NOT EXISTS comment_type VARCHAR(30) NOT NULL DEFAULT 'normal';

ALTER TABLE candidate_comments
DROP CONSTRAINT IF EXISTS chk_candidate_comments_type;

ALTER TABLE candidate_comments
ADD CONSTRAINT chk_candidate_comments_type
CHECK (comment_type IN ('normal', 'review_suitable', 'review_unsuitable'));

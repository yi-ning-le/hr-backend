-- Add created_by_user_id to interviews table to track which recruiter created the interview.
-- This is needed so we can notify the recruiter when the interview is completed.
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

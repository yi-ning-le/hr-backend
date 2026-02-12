-- Create candidate_comments table
CREATE TABLE IF NOT EXISTS candidate_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for retrieving comments for a candidate
CREATE INDEX IF NOT EXISTS idx_candidate_comments_candidate_id ON candidate_comments(candidate_id);

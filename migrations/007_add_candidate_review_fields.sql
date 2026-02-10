-- Add reviewer fields to candidates table
ALTER TABLE candidates
ADD COLUMN reviewer_id UUID REFERENCES employees(id),
ADD COLUMN review_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'suitable', 'unsuitable'
ADD COLUMN review_note TEXT;

-- Create index for filtering by reviewer
CREATE INDEX idx_candidates_reviewer_id ON candidates(reviewer_id);
CREATE INDEX idx_candidates_review_status ON candidates(review_status);

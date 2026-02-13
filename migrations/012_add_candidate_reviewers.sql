CREATE TABLE IF NOT EXISTS candidate_reviewers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id),
    reviewer_id UUID NOT NULL REFERENCES employees(id),
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    removed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(candidate_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_candidate_reviewers_reviewer ON candidate_reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_candidate_reviewers_candidate ON candidate_reviewers(candidate_id);

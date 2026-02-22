ALTER TABLE candidate_reviewers
ADD COLUMN IF NOT EXISTS review_status VARCHAR(20) NOT NULL DEFAULT 'pending',
ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;

ALTER TABLE candidate_reviewers
DROP CONSTRAINT IF EXISTS chk_candidate_reviewers_review_status;

ALTER TABLE candidate_reviewers
ADD CONSTRAINT chk_candidate_reviewers_review_status
CHECK (review_status IN ('pending', 'suitable', 'unsuitable'));

UPDATE candidate_reviewers cr
SET review_status = CASE
        WHEN c.review_status IN ('suitable', 'unsuitable') THEN c.review_status
        ELSE 'pending'
    END,
    reviewed_at = CASE
        WHEN c.review_status IN ('suitable', 'unsuitable') THEN c.updated_at
        ELSE NULL
    END
FROM candidates c
WHERE cr.candidate_id = c.id
  AND cr.reviewer_id = c.reviewer_id
  AND cr.removed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_candidate_reviewers_reviewer_status
ON candidate_reviewers (reviewer_id, review_status);

CREATE INDEX IF NOT EXISTS idx_candidate_reviewers_candidate_active
ON candidate_reviewers (candidate_id)
WHERE removed_at IS NULL;

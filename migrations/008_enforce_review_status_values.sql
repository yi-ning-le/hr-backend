ALTER TABLE candidates
DROP CONSTRAINT IF EXISTS chk_candidates_review_status;

ALTER TABLE candidates
ADD CONSTRAINT chk_candidates_review_status
CHECK (
  review_status IS NULL OR
  review_status IN ('pending', 'suitable', 'unsuitable')
);

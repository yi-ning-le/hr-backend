ALTER TABLE employees
ADD COLUMN IF NOT EXISTS can_review_resumes BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE employees e
SET can_review_resumes = TRUE
WHERE EXISTS (
  SELECT 1
  FROM interviews i
  WHERE i.interviewer_id = e.id
);

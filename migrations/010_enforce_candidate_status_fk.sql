UPDATE candidates c
SET status = 'new'
WHERE NOT EXISTS (
  SELECT 1
  FROM candidate_statuses cs
  WHERE cs.slug = c.status
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_candidates_status_slug'
  ) THEN
    ALTER TABLE candidates
    ADD CONSTRAINT fk_candidates_status_slug
    FOREIGN KEY (status)
    REFERENCES candidate_statuses(slug)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;
  END IF;
END $$;

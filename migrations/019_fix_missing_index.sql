-- Re-run cleanup to be safe
DELETE FROM interviews
WHERE id IN (
    SELECT id
    FROM (
        SELECT id,
               ROW_NUMBER() OVER (PARTITION BY candidate_id, job_id ORDER BY created_at DESC) as row_num
        FROM interviews
        WHERE status = 'PENDING'
    ) t
    WHERE t.row_num > 1
);

-- Ensure index exists (using IF NOT EXISTS to be safe)
CREATE UNIQUE INDEX IF NOT EXISTS idx_interviews_candidate_job_pending ON interviews(candidate_id, job_id) WHERE status = 'PENDING';

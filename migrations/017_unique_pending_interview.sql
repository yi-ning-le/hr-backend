-- 1. Clean up existing duplicate PENDING interviews (keep only the most recent one for each candidate per job)
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

-- 2. Create the partial unique index
CREATE UNIQUE INDEX idx_interviews_candidate_job_pending ON interviews(candidate_id, job_id) WHERE status = 'PENDING';

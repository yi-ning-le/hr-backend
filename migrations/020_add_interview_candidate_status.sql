ALTER TABLE interviews
ADD COLUMN candidate_status_id UUID REFERENCES candidate_statuses(id);

-- Backfill existing data
UPDATE interviews i
SET candidate_status_id = (
    SELECT s.id 
    FROM candidate_statuses s
    JOIN candidates c ON c.status = s.slug
    WHERE c.id = i.candidate_id
)
WHERE candidate_status_id IS NULL;

-- 1. Add snapshot fields to interviews table
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS snapshot_status_key VARCHAR(50);
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS snapshot_status_label VARCHAR(100);

-- 2. Migrate existing data based on candidate_status_id or candidate.status
-- First try to update from the linked candidate_status_id if available
UPDATE interviews i
SET 
    snapshot_status_key = s.slug,
    snapshot_status_label = s.name
FROM candidate_statuses s
WHERE i.candidate_status_id = s.id;

-- If still null (no candidate_status_id), try to infer from candidate current status
UPDATE interviews i
SET 
    snapshot_status_key = s.slug,
    snapshot_status_label = s.name
FROM candidate_statuses s
JOIN candidates c ON c.status = s.slug
WHERE i.snapshot_status_key IS NULL 
  AND i.candidate_id = c.id;

-- Fallback for any remaining nulls
UPDATE interviews
SET 
    snapshot_status_key = 'UNKNOWN',
    snapshot_status_label = 'Unknown Status'
WHERE snapshot_status_key IS NULL;

-- Enforce NOT NULL constraint after migration
ALTER TABLE interviews ALTER COLUMN snapshot_status_key SET NOT NULL;
ALTER TABLE interviews ALTER COLUMN snapshot_status_label SET NOT NULL;

-- 3. Add logical deletion to candidate_statuses
ALTER TABLE candidate_statuses ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- 4. Handle Constraints for Logical Deletion support
-- We MUST drop the Foreign Key on candidates.status because it relies on the strict unique constraint on slug.
-- By dropping the FK, we allow 'candidates' to hold status slugs that might be logically deleted (or duplicated in history).
ALTER TABLE candidates DROP CONSTRAINT IF EXISTS fk_candidates_status_slug;

-- Now we can drop the unique constraints on slug/name to allow reuse after logical deletion
ALTER TABLE candidate_statuses DROP CONSTRAINT IF EXISTS candidate_statuses_slug_key;
ALTER TABLE candidate_statuses DROP CONSTRAINT IF EXISTS candidate_statuses_name_key;

-- Create partial unique indexes (enforce uniqueness only for ACTIVE statuses)
-- This allows:
-- Row 1: slug='active', is_deleted=true
-- Row 2: slug='active', is_deleted=false  <-- Allowed
CREATE UNIQUE INDEX IF NOT EXISTS candidate_statuses_slug_active_idx ON candidate_statuses (slug) WHERE is_deleted = false;
CREATE UNIQUE INDEX IF NOT EXISTS candidate_statuses_name_active_idx ON candidate_statuses (name) WHERE is_deleted = false;

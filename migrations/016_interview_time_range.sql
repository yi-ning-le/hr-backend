ALTER TABLE interviews
ADD COLUMN IF NOT EXISTS scheduled_end_time TIMESTAMP WITH TIME ZONE;

UPDATE interviews
SET scheduled_end_time = scheduled_time + INTERVAL '1 hour'
WHERE scheduled_end_time IS NULL;

ALTER TABLE interviews
ALTER COLUMN scheduled_end_time SET NOT NULL;

ALTER TABLE interviews
DROP CONSTRAINT IF EXISTS interviews_scheduled_time_range_chk;

ALTER TABLE interviews
ADD CONSTRAINT interviews_scheduled_time_range_chk
CHECK (scheduled_end_time > scheduled_time);

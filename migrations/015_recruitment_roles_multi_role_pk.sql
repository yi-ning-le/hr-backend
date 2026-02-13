-- Allow an employee to hold multiple recruitment role types simultaneously.
-- Previous schema used employee_id as PK, which collapsed RECRUITER/INTERVIEWER into one slot.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'recruitment_roles_pkey'
    ) THEN
        ALTER TABLE recruitment_roles DROP CONSTRAINT recruitment_roles_pkey;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'recruitment_roles_pkey'
    ) THEN
        ALTER TABLE recruitment_roles
            ADD CONSTRAINT recruitment_roles_pkey PRIMARY KEY (employee_id, role_type);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_recruitment_roles_employee_id
    ON recruitment_roles(employee_id);

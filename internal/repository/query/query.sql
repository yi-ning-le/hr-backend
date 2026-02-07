-- name: CreateJob :one
INSERT INTO jobs (
  title, department, head_count, open_date, job_description, note, status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs
WHERE id = $1 LIMIT 1;

-- name: ListJobs :many
SELECT * FROM jobs
ORDER BY created_at DESC;

-- name: UpdateJob :one
UPDATE jobs
SET title = $2,
    department = $3,
    head_count = $4,
    open_date = $5,
    job_description = $6,
    note = $7,
    status = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateJobStatus :one
UPDATE jobs
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteJob :exec
DELETE FROM jobs
WHERE id = $1;

-- name: CreateCandidate :one
INSERT INTO candidates (
  name, avatar, email, phone, experience_years, education, applied_job_id, channel, resume_url, status, note, applied_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetCandidate :one
SELECT c.*, j.title as applied_job_title
FROM candidates c
JOIN jobs j ON c.applied_job_id = j.id
WHERE c.id = $1 LIMIT 1;

-- name: ListCandidates :many
SELECT c.*, j.title as applied_job_title
FROM candidates c
JOIN jobs j ON c.applied_job_id = j.id
WHERE ($1::uuid IS NULL OR c.applied_job_id = $1)
ORDER BY c.applied_at DESC;

-- name: UpdateCandidate :one
UPDATE candidates
SET name = $2,
    avatar = $3,
    email = $4,
    phone = $5,
    experience_years = $6,
    education = $7,
    applied_job_id = $8,
    channel = $9,
    resume_url = $10,
    status = $11,
    note = $12,
    applied_at = $13,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateCandidateStatus :one
UPDATE candidates
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateCandidateNote :one
UPDATE candidates
SET note = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateCandidateResume :one
UPDATE candidates
SET resume_url = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteCandidate :exec
DELETE FROM candidates
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, avatar)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- Employee queries

-- name: CreateEmployee :one
INSERT INTO employees (
  first_name, last_name, email, phone, department, position, status, employment_type, join_date, manager_id, user_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetEmployee :one
SELECT * FROM employees
WHERE id = $1 LIMIT 1;

-- name: ListEmployees :many
SELECT * FROM employees
WHERE (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('department')::varchar IS NULL OR department = sqlc.narg('department'))
  AND (sqlc.narg('search')::varchar IS NULL OR first_name ILIKE '%' || sqlc.narg('search') || '%' OR last_name ILIKE '%' || sqlc.narg('search') || '%' OR email ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY created_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees
WHERE (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('department')::varchar IS NULL OR department = sqlc.narg('department'))
  AND (sqlc.narg('search')::varchar IS NULL OR first_name ILIKE '%' || sqlc.narg('search') || '%' OR last_name ILIKE '%' || sqlc.narg('search') || '%' OR email ILIKE '%' || sqlc.narg('search') || '%');

-- name: UpdateEmployee :one
UPDATE employees
SET first_name = $2,
    last_name = $3,
    email = $4,
    phone = $5,
    department = $6,
    position = $7,
    status = $8,
    employment_type = $9,
    join_date = $10,
    manager_id = $11,
    user_id = $12,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteEmployee :exec
DELETE FROM employees
WHERE id = $1;

-- Candidate Status queries

-- name: ListCandidateStatuses :many
SELECT * FROM candidate_statuses
ORDER BY sort_order ASC;

-- name: GetCandidateStatus :one
SELECT * FROM candidate_statuses
WHERE id = $1 LIMIT 1;

-- name: GetCandidateStatusBySlug :one
SELECT * FROM candidate_statuses
WHERE slug = $1 LIMIT 1;

-- name: CreateCandidateStatus :one
INSERT INTO candidate_statuses (
    name, slug, type, sort_order, color
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateCandidateStatusFields :one
UPDATE candidate_statuses
SET name = $2,
    color = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateCandidateStatusOrder :exec
UPDATE candidate_statuses
SET sort_order = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteCandidateStatus :exec
DELETE FROM candidate_statuses
WHERE id = $1;

-- Recruitment Role queries

-- name: CheckIsAdmin :one
SELECT is_admin FROM users WHERE id = $1 LIMIT 1;

-- name: CheckRecruiterRole :one
SELECT employee_id FROM recruitment_roles WHERE employee_id = $1 LIMIT 1;

-- name: GetActiveInterviewCount :one
SELECT COUNT(*) FROM interviews 
WHERE interviewer_id = $1 AND status = 'PENDING';

-- name: AssignRecruiterRole :exec
INSERT INTO recruitment_roles (employee_id, role_type)
VALUES ($1, 'RECRUITER')
ON CONFLICT (employee_id) DO NOTHING;

-- name: RevokeRecruiterRole :exec
DELETE FROM recruitment_roles WHERE employee_id = $1;

-- name: ListRecruiters :many
SELECT e.id, e.first_name, e.last_name, e.department, e.phone
FROM recruitment_roles rr
JOIN employees e ON rr.employee_id = e.id
ORDER BY e.first_name;

-- name: GetEmployeeByUserID :one
SELECT * FROM employees WHERE user_id = $1 LIMIT 1;

-- Interview queries

-- name: CreateInterview :one
INSERT INTO interviews (
  candidate_id, interviewer_id, job_id, scheduled_time, status, notes
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetInterview :one
SELECT * FROM interviews WHERE id = $1 LIMIT 1;

-- name: ListInterviewsByInterviewer :many
SELECT * FROM interviews
WHERE interviewer_id = $1
ORDER BY scheduled_time DESC;

-- name: TransferInterview :one
UPDATE interviews
SET interviewer_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateInterviewStatus :one
UPDATE interviews
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
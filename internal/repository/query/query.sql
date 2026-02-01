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
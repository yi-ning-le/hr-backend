-- name: GetNotificationsByUserId :many
SELECT id, user_id, event_type, subject_type, subject_id, context, read_at, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*)
FROM notifications
WHERE user_id = $1 AND read_at IS NULL;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET read_at = COALESCE(read_at, CURRENT_TIMESTAMP)
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET read_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND read_at IS NULL;

-- name: CreateNotification :one
INSERT INTO notifications (
    user_id,
    event_type,
    subject_type,
    subject_id,
    context
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, user_id, event_type, subject_type, subject_id, context, read_at, created_at;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: DeleteNotificationsBySubjectAndType :exec
DELETE FROM notifications
WHERE user_id = $1
  AND subject_type = $2
  AND subject_id = $3
  AND event_type = $4;

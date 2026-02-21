-- name: GetNotificationsByUserId :many
SELECT *
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*)
FROM notifications
WHERE user_id = $1 AND is_read = false;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET is_read = true
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET is_read = true
WHERE user_id = $1 AND is_read = false;

-- name: CreateNotification :one
INSERT INTO notifications (
    user_id,
    title,
    message,
    type,
    link_url
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: CreateReminder :one
INSERT INTO reminders (
  user_id,
  website_url,
  interval,
  extension
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: DeleteReminder :exec
DELETE FROM reminders
WHERE id = $1 AND website_url = $2;

-- name: SetNewInterval :one
UPDATE reminders
SET interval = sqlc.arg(new_interval)
WHERE id = sqlc.arg(id) AND website_url = sqlc.arg(website_url)
RETURNING *;

-- name: UpdateReminder :one
UPDATE reminders
SET updated_at = $1
WHERE id = $2 AND website_url = $3
RETURNING *;

-- name: GetReminderConfigs :one
SELECT extension FROM reminders
WHERE id = $1 AND website_url = $2
FOR NO KEY UPDATE;

-- name: SetReminderConfigs :one
UPDATE reminders
SET extension = sqlc.arg(updated_extension)
WHERE id = sqlc.arg(id) AND website_url = sqlc.arg(website_url)
RETURNING *;

-- name: GetReminder :one
SELECT * FROM reminders
WHERE id = $1 AND website_url = $2
LIMIT 1;

-- name: ListReminders :many
SELECT * FROM reminders
WHERE user_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;
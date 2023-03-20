-- name: CreateUser :one
INSERT INTO users (
  full_name,
  hashed_password,
  email
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = sqlc.arg(user_id) LIMIT 1;

-- name: DeactivateUser :one
UPDATE users
SET is_activated = false
WHERE id = $1 AND email = $2
RETURNING *;

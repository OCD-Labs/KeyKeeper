-- name: CreateUser :one
INSERT INTO users (
  full_name,
  hashed_password,
  email,
  profile_image_url
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = sqlc.arg(user_id) LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = sqlc.arg(user_email) LIMIT 1;

-- name: DeactivateUser :one
UPDATE users
SET is_active = false
WHERE id = $1 AND email = $2
RETURNING *;

-- name: VerifyEmail :one
UPDATE users
SET is_email_verified = true, is_active = true
WHERE email = $1
RETURNING *;

-- name: ActivateUser :one
UPDATE users
SET is_active = true
WHERE email = $1
RETURNING *;

-- name: ChangePassword :one
UPDATE users
SET hashed_password = $1, password_changed_at = now()
WHERE id = $2
RETURNING *;

-- name: ChangeEmail :one
UPDATE users
SET email = $1
WHERE id = $2
RETURNING *;

-- name: ChangeProfileImage :one
UPDATE users
SET profile_image_url = $1
WHERE id = $2
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET
  hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
  password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at),
  full_name = COALESCE(sqlc.narg(full_name), full_name),
  email = COALESCE(sqlc.narg(email), email),
  is_email_verified = COALESCE(sqlc.narg(is_email_verified), is_email_verified),
  is_active = COALESCE(sqlc.narg(is_active), is_active),
  profile_image_url = COALESCE(sqlc.narg(profile_image_url), profile_image_url)
WHERE 
  id = sqlc.narg(id) OR email = sqlc.narg(email)
RETURNING *;

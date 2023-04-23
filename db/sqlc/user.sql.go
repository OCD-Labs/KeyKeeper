// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: user.sql

package db

import (
	"context"
	"database/sql"
)

const changeEmail = `-- name: ChangeEmail :one
UPDATE users
SET email = $1
WHERE id = $2
RETURNING id, full_name, hashed_password, email, profile_image_url, password_changed_at, created_at, is_active, is_email_verified
`

type ChangeEmailParams struct {
	Email string `json:"email"`
	ID    int64  `json:"id"`
}

func (q *Queries) ChangeEmail(ctx context.Context, arg ChangeEmailParams) (User, error) {
	row := q.db.QueryRowContext(ctx, changeEmail, arg.Email, arg.ID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FullName,
		&i.HashedPassword,
		&i.Email,
		&i.ProfileImageUrl,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsActive,
		&i.IsEmailVerified,
	)
	return i, err
}

const changePassword = `-- name: ChangePassword :one
UPDATE users
SET hashed_password = $1
WHERE email = $2
RETURNING id, full_name, hashed_password, email, profile_image_url, password_changed_at, created_at, is_active, is_email_verified
`

type ChangePasswordParams struct {
	HashedPassword string `json:"hashed_password"`
	Email          string `json:"email"`
}

func (q *Queries) ChangePassword(ctx context.Context, arg ChangePasswordParams) (User, error) {
	row := q.db.QueryRowContext(ctx, changePassword, arg.HashedPassword, arg.Email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FullName,
		&i.HashedPassword,
		&i.Email,
		&i.ProfileImageUrl,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsActive,
		&i.IsEmailVerified,
	)
	return i, err
}

const changeProfileImage = `-- name: ChangeProfileImage :one
UPDATE users
SET profile_image_url = $1
WHERE id = $2
RETURNING id, full_name, hashed_password, email, profile_image_url, password_changed_at, created_at, is_active, is_email_verified
`

type ChangeProfileImageParams struct {
	ProfileImageUrl sql.NullString `json:"profile_image_url"`
	ID              int64          `json:"id"`
}

func (q *Queries) ChangeProfileImage(ctx context.Context, arg ChangeProfileImageParams) (User, error) {
	row := q.db.QueryRowContext(ctx, changeProfileImage, arg.ProfileImageUrl, arg.ID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FullName,
		&i.HashedPassword,
		&i.Email,
		&i.ProfileImageUrl,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsActive,
		&i.IsEmailVerified,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  full_name,
  hashed_password,
  email,
  profile_image_url
) VALUES (
  $1, $2, $3, $4
) RETURNING id, full_name, hashed_password, email, profile_image_url, password_changed_at, created_at, is_active, is_email_verified
`

type CreateUserParams struct {
	FullName        string         `json:"full_name"`
	HashedPassword  string         `json:"hashed_password"`
	Email           string         `json:"email"`
	ProfileImageUrl sql.NullString `json:"profile_image_url"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.FullName,
		arg.HashedPassword,
		arg.Email,
		arg.ProfileImageUrl,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FullName,
		&i.HashedPassword,
		&i.Email,
		&i.ProfileImageUrl,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsActive,
		&i.IsEmailVerified,
	)
	return i, err
}

const deactivateUser = `-- name: DeactivateUser :one
UPDATE users
SET is_active = false
WHERE id = $1 AND email = $2
RETURNING id, full_name, hashed_password, email, profile_image_url, password_changed_at, created_at, is_active, is_email_verified
`

type DeactivateUserParams struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func (q *Queries) DeactivateUser(ctx context.Context, arg DeactivateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, deactivateUser, arg.ID, arg.Email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FullName,
		&i.HashedPassword,
		&i.Email,
		&i.ProfileImageUrl,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsActive,
		&i.IsEmailVerified,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT id, full_name, hashed_password, email, profile_image_url, password_changed_at, created_at, is_active, is_email_verified FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, userID int64) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, userID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FullName,
		&i.HashedPassword,
		&i.Email,
		&i.ProfileImageUrl,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.IsActive,
		&i.IsEmailVerified,
	)
	return i, err
}

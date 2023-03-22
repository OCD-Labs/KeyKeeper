// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	ID         int64           `json:"id"`
	UserID     int64           `json:"user_id"`
	WebsiteUrl string          `json:"website_url"`
	Interval   string          `json:"interval"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Extension  json.RawMessage `json:"extension"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID                int64     `json:"id"`
	FullName          string    `json:"full_name"`
	HashedPassword    string    `json:"hashed_password"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	IsActivated       bool      `json:"is_activated"`
}

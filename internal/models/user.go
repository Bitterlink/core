// Package models is used to store data models
package models

import (
	"database/sql" // Required for sql.NullString and sql.NullTime
	"time"         // Required for time.Time
)

// User represents a user account in the BitterLink system.
// It maps to the `users` table in the database.
type User struct {
	// ID corresponds to the `id` column (BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY).
	// int64 is the standard Go type for database IDs.
	ID int64 `json:"id"`

	// Name corresponds to the `name` column (VARCHAR(255) NULL).
	// sql.NullString handles nullable VARCHAR fields correctly.
	Name string `json:"name"`

	// Email corresponds to the `email` column (VARCHAR(255) UNIQUE NOT NULL).
	Email string `json:"email"`

	// PasswordHash corresponds to the `password_hash` column (VARCHAR(255) NOT NULL).
	// It stores the securely hashed password.
	// Tag `json:"-"` prevents this field from ever being included in JSON responses.
	PasswordHash string `json:"-"`

	// EmailVerifiedAt corresponds to the `email_verified_at` column (TIMESTAMP NULL).
	// sql.NullTime handles nullable TIMESTAMP/DATETIME fields.
	// omitempty tag means it won't appear in JSON if it's null/zero.
	EmailVerifiedAt sql.NullTime `json:"email_verified_at,omitempty"`

	// RememberToken corresponds to the `remember_token` column (VARCHAR(100) NULL).
	// Often used by web frameworks like Laravel. Excluded from JSON.
	RememberToken sql.NullString `json:"-"`

	// DeletedAt corresponds to the `deleted_at` column (TIMESTAMP NULL).
	// Used for soft deletes. Excluded from standard JSON responses.
	DeletedAt sql.NullTime `json:"-"`

	// CreatedAt corresponds to the `created_at` column (TIMESTAMP NULL DEFAULT...).
	// Assumes `parseTime=true` in your DSN, mapping TIMESTAMP to time.Time.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt corresponds to the `updated_at` column (TIMESTAMP NULL DEFAULT...).
	UpdatedAt time.Time `json:"updated_at"`
}

// --- Helper Methods (Optional but Recommended) ---

// IsVerified checks if the EmailVerifiedAt timestamp is set (meaning not NULL).
func (u *User) IsVerified() bool {
	return u.EmailVerifiedAt.Valid
}

// IsDeleted checks if the DeletedAt timestamp is set (meaning not NULL).
func (u *User) IsDeleted() bool {
	return u.DeletedAt.Valid
}

// GetName returns the user's name, or an empty string if the name is NULL.
func (u *User) GetName() string {
	return u.Name
}

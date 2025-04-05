package models

import (
	"database/sql"
	"time"
)

// Check represents the data structure for a monitored check.
type Check struct {
	ID               int64          `json:"id"`
	UserID           int64          `json:"user_id"` // Or omit from JSON if not needed client-side
	UUID             string         `json:"uuid"`    // Public ID
	Name             string         `json:"name"`
	Description      sql.NullString `json:"description"`       // Handles NULL TEXT
	ExpectedInterval uint32         `json:"expected_interval"` // Assuming INT UNSIGNED
	GracePeriod      uint32         `json:"grace_period"`      // Assuming INT UNSIGNED
	LastPingAt       sql.NullTime   `json:"last_ping_at"`      // Handles NULL TIMESTAMP
	Status           string         `json:"status"`            // ENUM maps nicely to string
	IsEnabled        bool           `json:"is_enabled"`
	CreatedAt        time.Time      `json:"created_at"` // Assumes parseTime=True in DSN
	UpdatedAt        time.Time      `json:"updated_at"`
}

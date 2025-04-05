package repository

import (
	"bitterlink/core/internal/models"
	"context"
	"database/sql"
)

type CheckRepository interface {
	FindByID(ctx context.Context, id int64) (*models.Check, error)
	FindByUUID(ctx context.Context, uuid string) (*models.Check, error)
	FindActiveByUserID(ctx context.Context, userID int64) ([]models.Check, error) // Like our previous example!
	Create(ctx context.Context, check *models.Check) error                        // Might return the ID or the full check
	Update(ctx context.Context, check *models.Check) error
	Delete(ctx context.Context, id int64) error                                                           // Handles soft delete logic
	RecordPing(ctx context.Context, uuid string, sourceIP sql.NullString, userAgent sql.NullString) error // Added sourceIP/userAgent
	ListByUserID(ctx context.Context, userID int64) ([]models.Check, error)
	// ... other methods as needed (e.g., UpdateStatus, UpdateLastPing)
}

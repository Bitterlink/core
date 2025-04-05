package repository // Or handlers, datastore, etc.

import (
	"context" // Always pass context for cancellation/timeouts
	"database/sql"
	"errors"
	"fmt" // For error wrapping
	"log"

	"bitterlink/core/internal/models" // Import your Check struct definition

	"github.com/go-sql-driver/mysql"
)

// ErrCheckNotFound --- Add Custom Error ---
var ErrCheckNotFound = errors.New("check not found or not active")

// Create inserts a new Check record into the database.
// It sets the auto-generated ID and potentially CreatedAt/UpdatedAt
// back onto the input check pointer upon success.
func (r *mysqlCheckRepository) Create(ctx context.Context, check *models.Check) error {
	// 1. Basic Validation (more complex validation often belongs in a service layer)
	if check == nil {
		return errors.New("can not create nil check")
	}
	if check.UserID <= 0 {
		return errors.New("UserID is required to create a check")
	}
	if check.UUID == "" {
		// The application layer (e.g., handler or service) should generate
		// the UUID before calling the repository.
		return errors.New("UUID is required to create a check")
	}
	if check.Name == "" {
		return errors.New("Name is required to create a check")
	}
	if check.ExpectedInterval <= 0 {
		return errors.New("ExpectedInterval must be greater than zero")
	}

	// 2. Define the INSERT Query
	// We specify the columns we are providing values for.
	// Let the DB handle defaults for id, last_ping_at, deleted_at,
	// but explicitly set created_at and updated_at using UTC_TIMESTAMP().
	query := `
        INSERT INTO checks (
            user_id, uuid, name, description, expected_interval, grace_period,
            status, is_enabled, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, UTC_TIMESTAMP(), UTC_TIMESTAMP())`

	// 3. Prepare Arguments
	// Set default status if empty (e.g., 'new')
	status := check.Status
	if status == "" {
		status = "new"
	}
	// Set default enabled state (usually true)
	isEnabled := true // Assuming default is false if not set, make it true?
	// Let's assume new checks are enabled unless specified otherwise.
	// If check.IsEnabled was not explicitly set before calling repo, default it.
	// However, it's better if the caller sets IsEnabled explicitly. For safety:
	// isEnabled := true // Or use check.IsEnabled if the caller sets it.

	// 4. Execute the Query
	// Use ExecContext for INSERT, UPDATE, DELETE statements.
	result, err := r.db.ExecContext(
		ctx,
		query,
		check.UserID,
		check.UUID,
		check.Name,
		check.Description, // Pass sql.NullString directly
		check.ExpectedInterval,
		check.GracePeriod,
		status,    // Use the determined status
		isEnabled, // Use the value from the struct (caller should set default)
	)

	// 5. Handle Errors
	if err != nil {
		// Check for specific MySQL errors, like duplicate entry for UNIQUE constraints
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 { // 1062 is 'Duplicate entry'
			// Could check mysqlErr.Message to see which key constraint failed (e.g., uuid)
			log.Printf("WARN: Attempted to create check with duplicate entry (likely UUID '%s'): %v", check.UUID, err)
			return fmt.Errorf("check with this UUID already exists: %w", err)
		}
		// Log generic database error
		log.Printf("ERROR: Failed to insert check for user %d (UUID: %s): %v", check.UserID, check.UUID, err)
		return fmt.Errorf("database error creating check: %w", err)
	}

	// 6. Get the Last Inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		// This is less likely but possible
		log.Printf("ERROR: Failed to get last insert ID for check UUID %s: %v", check.UUID, err)
		// The insert likely succeeded, but we can't confirm the ID. Critical? Maybe return error.
		return fmt.Errorf("failed to retrieve new check ID after insert: %w", err)
	}

	// 7. Update the input struct pointer with the new ID
	check.ID = id
	// We could also set check.CreatedAt/UpdatedAt based on time.Now(), but the DB values are the source of truth.
	// Setting the ID is usually the most important part.
	check.Status = status // Ensure status is set if defaulted

	log.Printf("INFO: Successfully created check with ID %d (UUID: %s)", check.ID, check.UUID)
	return nil // Success!
}

func (r *mysqlCheckRepository) Update(ctx context.Context, check *models.Check) error {
	// TODO: Implement SQL UPDATE statement using r.db.ExecContext
	log.Printf("DEBUG: Update check called (Not Implemented): ID=%d", check.ID)
	return fmt.Errorf("repository Update method not implemented yet")
}

func (r *mysqlCheckRepository) Delete(ctx context.Context, id int64) error {
	// TODO: Implement soft delete (UPDATE checks SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL)
	log.Printf("DEBUG: Delete check called (Not Implemented): ID=%d", id)
	return fmt.Errorf("repository Delete method not implemented yet")
}

// FindByID Add FindByID if you haven't already
func (r *mysqlCheckRepository) FindByID(ctx context.Context, id int64) (*models.Check, error) {
	// TODO: Implement SQL SELECT ... WHERE id = ? AND deleted_at IS NULL logic using r.db.QueryRowContext
	log.Printf("DEBUG: FindByID check called (Not Implemented): ID=%d", id)
	// Example of returning ErrCheckNotFound if appropriate
	// return nil, ErrCheckNotFound
	return nil, fmt.Errorf("repository FindByID method not implemented yet")
}

// FindActiveByUserID Ensure FindActiveByUserID is also implemented if it's in the interface
func (r *mysqlCheckRepository) FindActiveByUserID(ctx context.Context, userID int64) ([]models.Check, error) {
	// TODO: Implement the logic from the previous example if you haven't moved it here yet
	log.Printf("DEBUG: FindActiveByUserID check called (Not Implemented): UserID=%d", userID)
	return nil, fmt.Errorf("repository FindActiveByUserID method not implemented yet")
}

// mysqlCheckRepository implements CheckRepository using a MySQL database
type mysqlCheckRepository struct {
	db *sql.DB
}

// NewMySQLCheckRepository creates a new repository instance
func NewMySQLCheckRepository(dbPool *sql.DB) CheckRepository {
	return &mysqlCheckRepository{db: dbPool}
}

// RecordPing --- Implement RecordPing ---
// RecordPing finds a check by UUID, updates its last ping time and status (if down),
// and inserts a record into the pings table. It performs these operations in a transaction.
func (r *mysqlCheckRepository) RecordPing(ctx context.Context, uuid string, sourceIP sql.NullString, userAgent sql.NullString) error {
	// Use a transaction to ensure atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var checkID int64
	var currentStatus string
	findQuery := "SELECT id, status FROM checks WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err = tx.QueryRowContext(ctx, findQuery, uuid).Scan(&checkID, &currentStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Use the custom error for clear handling in the handler
			return ErrCheckNotFound
		}
		// Log the technical error but return a generic one potentially
		log.Printf("ERROR: RecordPing - Failed to find check by UUID '%s': %v", uuid, err)
		return fmt.Errorf("database error finding check: %w", err)
	}

	// 2. Update the check's last_ping_at and status (if it was 'down')
	// Note: We update last_ping_at even for 'paused' checks, but status only flips from 'down'.
	// If it's a new check. The first ping brings it to up.
	newStatus := currentStatus
	if currentStatus == "down" || currentStatus == "new" {
		newStatus = "up"
	}

	updateQuery := `
        UPDATE checks
        SET last_ping_at = UTC_TIMESTAMP(), status = ?, updated_at = UTC_TIMESTAMP()
        WHERE id = ?`
	_, err = tx.ExecContext(ctx, updateQuery, newStatus, checkID)
	if err != nil {
		log.Printf("ERROR: RecordPing - Failed to update check ID %d: %v", checkID, err)
		return fmt.Errorf("database error updating check: %w", err)
	}

	// 3. Insert the ping details into the pings table
	// For now, payload is NULL. Handle payload later if needed.
	insertQuery := `
        INSERT INTO pings (check_id, received_at, source_ip, user_agent, payload, created_at)
        VALUES (?, UTC_TIMESTAMP(), ?, ?, NULL, UTC_TIMESTAMP())`
	_, err = tx.ExecContext(ctx, insertQuery, checkID, sourceIP, userAgent)
	if err != nil {
		log.Printf("ERROR: RecordPing - Failed to insert ping record for check ID %d: %v", checkID, err)
		return fmt.Errorf("database error recording ping details: %w", err)
	}

	// 4. If all went well, commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("ERROR: RecordPing - Failed to commit transaction for check ID %d: %v", checkID, err)
		return fmt.Errorf("database error committing ping record: %w", err)
	}

	log.Printf("DEBUG: Successfully recorded ping for check ID %d (UUID: %s)", checkID, uuid)
	return nil // Success

}

// FindByUUID Implement other CheckRepository methods (FindByID, Create, etc.) here...
// Example: FindByUUID (useful for other parts of the API perhaps)
func (r *mysqlCheckRepository) FindByUUID(ctx context.Context, uuid string) (*models.Check, error) {
	query := `SELECT id, user_id, uuid, name, description, expected_interval, grace_period, 
                     last_ping_at, status, is_enabled, created_at, updated_at 
              FROM checks WHERE uuid = ? AND deleted_at IS NULL LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, uuid)
	var check models.Check
	err := row.Scan(
		&check.ID, &check.UserID, &check.UUID, &check.Name, &check.Description,
		&check.ExpectedInterval, &check.GracePeriod, &check.LastPingAt, &check.Status,
		&check.IsEnabled, &check.CreatedAt, &check.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCheckNotFound
		}
		log.Printf("ERROR: FindByUUID - Scan failed for UUID %s: %v", uuid, err)
		return nil, fmt.Errorf("error retrieving check data: %w", err)
	}
	return &check, nil
}

// ListByUserID GetActiveChecksForUser retrieves all non-deleted checks for a specific user.
func (r *mysqlCheckRepository) ListByUserID(ctx context.Context, userID int64) ([]models.Check, error) {

	// 1. Define the SQL Query
	// Select the columns in the order you expect to Scan them.
	// Filter by user_id and make sure deleted_at IS NULL for soft delete.
	query := `
		SELECT
			id, user_id, uuid, name, description, expected_interval,
			grace_period, last_ping_at, status, is_enabled, created_at, updated_at
		FROM checks
		WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY name ASC` // Or ORDER BY created_at, etc.

	// 2. Execute the Query using QueryContext
	// Pass the context, query string, and any arguments (userID in this case).
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("ERROR: dbPool.QueryContext failed for user %d: %v", userID, err)
		// Return a wrapped error for context, hiding internal details if necessary
		return nil, fmt.Errorf("error querying user checks: %w", err)
	}
	// 3. IMPORTANT: Ensure rows is closed eventually to return the connection
	// Defer guarantees it runs even if errors occur during scanning.
	defer rows.Close()

	// 4. Prepare to collect the results
	var checks []models.Check // Initialize an empty slice

	// 5. Iterate through the result set
	for rows.Next() { // .Next() prepares the next row for reading
		var check models.Check // Create a Check struct to scan data into

		// 6. Scan the values from the current row into the Check struct fields
		// The order of &check.Field must EXACTLY match the order of columns in SELECT.
		err := rows.Scan(
			&check.ID,
			&check.UserID,
			&check.UUID,
			&check.Name,
			&check.Description, // Scan directly into sql.NullString
			&check.ExpectedInterval,
			&check.GracePeriod,
			&check.LastPingAt, // Scan directly into sql.NullTime
			&check.Status,
			&check.IsEnabled,
			&check.CreatedAt,
			&check.UpdatedAt,
		)
		if err != nil {
			// Log the error and potentially stop processing, returning the error.
			log.Printf("ERROR: Failed to scan row for user %d check: %v", userID, err)
			return nil, fmt.Errorf("error scanning check data: %w", err)
		}

		// 7. Append the successfully scanned check to the results slice
		checks = append(checks, check)
	}

	// 8. Check for errors that may have occurred during iteration
	if err = rows.Err(); err != nil {
		log.Printf("ERROR: Error during row iteration for user %d checks: %v", userID, err)
		return nil, fmt.Errorf("error iterating check results: %w", err)
	}

	// 9. Return the results (checks will be an empty slice if no rows found, not nil)
	log.Printf("INFO: Found %d checks for user %d", len(checks), userID)
	return checks, nil
}

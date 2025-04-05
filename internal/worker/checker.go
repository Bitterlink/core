// Package to handle checks
package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Config struct {
	PollInterval time.Duration
	BatchSize int
}

type TimeoutChecker struct {
	dbPool *sql.DB
	config Config
	// Add a message queue producer here later for notifications
	// notificationDispatcher NotificationDispatcher // Example interface
}

// NewTimeoutChecker creates a new checker instance.
func NewTimeoutChecker(db *sql.DB, cfg Config) *TimeoutChecker {
	return &TimeoutChecker{
		dbPool: db,
		config: cfg,
	}
}

// Start runs the periodic check loop until the context is cancelled.
func (tc *TimeoutChecker) Start(ctx context.Context) {
	log.Printf("INFO: Starting TimeoutChecker worker with poll interval %v", tc.config.PollInterval)
	// Create a ticker that fires at the configured interval
	ticker := time.NewTicker(tc.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Time to check for timeouts
			log.Println("DEBUG: TimeoutChecker tick: processing timeouts...")
			err := tc.processTimeouts(ctx)
			if err != nil {
				// Log the error but continue running
				log.Printf("ERROR: Error processing timeouts: %v", err)
			}
		case <-ctx.Done():
			// Context was cancelled (e.g., shutdown signal)
			log.Println("INFO: TimeoutChecker worker stopping due to context cancellation.")
			return // Exit the loop and the goroutine
		}
	}
}

func (tc *TimeoutChecker) processTimeouts(ctx context.Context) error {
	// 1. Begin Transaction
	tx, err := tc.dbPool.BeginTx(ctx, nil) // Use default isolation level 
	if err != nil {
		return fmt.Errorf("failed to begin transation: %w", err)
	}
	defer tx.Rollback()

	// 2. Execute Query to Find and Lock Timed-out Checks
	// Using UTC_TIMESTAMP() for database time comparison is generally safer
	query := `
        SELECT id, uuid -- Select minimal info needed to process/notify
        FROM checks
        WHERE
            status = 'up'
            AND is_enabled = TRUE
            AND deleted_at IS NULL
            AND last_ping_at < (UTC_TIMESTAMP() - INTERVAL (expected_interval + grace_period) SECOND)
        ORDER BY last_ping_at ASC -- Process oldest first
        LIMIT ? -- Use configured batch size
        FOR UPDATE SKIP LOCKED` // The key part for concurrency

	rows, err := tx.QueryContext(ctx, query, tc.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to query timed-out checks: %w", err)
	}
	defer rows.Close()

	var checkIDsToProcess []int64
	var timedOutChecksInfo []string // for logging

	// 3. Collect IDs of checks to process
	for rows.Next() {
		var id int64
		var uuid string
		if err := rows.Scan(&id, &uuid); err != nil {
			// Log error but potentially continue processing others found so far?
			// For simplicity, let's return error and rollback the whole batch on scan failure.
			return fmt.Errorf("failed to scan check row: %w", err) 
		}
		checkIDsToProcess = append(checkIDsToProcess, id)
		timedOutChecksInfo = append(timedOutChecksInfo, fmt.Sprintf("%d (%s)", id, uuid))
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("row iteration failed: %w", err) 
	}

	// If no rows found, commit the empty transaction and exit successfully
	if len(checkIDsToProcess) == 0 {
		return tx.Commit() // Commit needed even if empty to finish tx
	}

	log.Printf("INFO: Found %d timed-out checks to process: %v", len(checkIDsToProcess), timedOutChecksInfo)

	// 4. Process Locked Rows (Update Status & Dispatch Notifications)
	updateQuery := `UPDATE checks SET status = 'down', updated_at = UTC_TIMESTAMP() WHERE id = ?`
	for _, checkID := range checkIDsToProcess {
		// Update status within the same transaction
		_, updateErr := tx.ExecContext(ctx, updateQuery, checkID)
		if updateErr != nil {
			// Rollback will happen via defer
			return fmt.Errorf("failed to update status for check ID %d: %w", checkID, updateErr)
		}
		log.Printf("DEBUG: Marked check ID %d as down.", checkID)

		// !!! TODO: Dispatch notification task HERE !!!
		// Example: tc.notificationDispatcher.DispatchDownNotification(ctx, checkID)
		// This should ideally send a message (with checkID/UUID/UserID)
		// to a message queue (like RabbitMQ/Redis) for a separate worker
		// to handle the actual sending of email/Slack/webhook.
		log.Printf("INFO: Dispatched 'down' notification task for check ID %d", checkID)
	}

	// 5. Commit Transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("INFO: Successfully processed batch of %d timed-out checks.", len(checkIDsToProcess))
	return nil
}
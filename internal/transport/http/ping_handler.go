package httptransport

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"bitterlink/core/internal/repository"

	"github.com/gin-gonic/gin"
)

// PingHandler holds dependencies for ping routes
type PingHandler struct {
	CheckRepo repository.CheckRepository
}

// NewPingHandler creates a new handler for ping operations
func NewPingHandler(cr repository.CheckRepository) *PingHandler {
	return &PingHandler{
		CheckRepo: cr,
	}
}

// HandlePing processes incoming pings for a check identified by UUID.
// Method: GET or POST /ping/{uuid}
func (h *PingHandler) HandlePing(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		// Although route matching usually prevents this, good to check.
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing check UUID parameter"})
		return
	}

	// Optional: Validate UUID format if desired
	// e.g., using a regex or a UUID library

	// Capture client info (handle potential nulls for DB)
	clientIP := sql.NullString{
		String: c.ClientIP(),
		Valid:  c.ClientIP() != "",
	}
	userAgent := sql.NullString{
		String: c.Request.UserAgent(),
		Valid:  c.Request.UserAgent() != "",
	}
	ctx := c.Request.Context() // Use request context

	err := h.CheckRepo.RecordPing(ctx, uuid, clientIP, userAgent)

	if err != nil {
		// Check for the specific "not found" error from the repository
		if errors.Is(err, repository.ErrCheckNotFound) {
			log.Printf("WARN: Ping received for unknown/inactive UUID: %s", uuid)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Check not found or inactive"})
		} else {
			// Log the underlying error details for server-side debugging
			log.Printf("ERROR: Failed processing ping for UUID %s: %v", uuid, err)
			// Return a generic server error to the client
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process ping"})
		}
		return // Stop processing
	}

	// Success!
	// Return a simple 'ok' response.
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

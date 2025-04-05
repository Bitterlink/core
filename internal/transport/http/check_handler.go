package httptransport

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"bitterlink/core/internal/middleware"
	"bitterlink/core/internal/models"
	"bitterlink/core/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateCheckRequest struct {
	Name             string  `json:"name" binding:"required"`                   // Use Gin binding tags for validation
	Description      *string `json:"description"`                               // Pointer handles null/omitted vs ""
	ExpectedInterval uint32  `json:"expected_interval" binding:"required,gt=0"` // required, greater than 0
	GracePeriod      *uint32 `json:"grace_period"`                              // Pointer handles null/omitted vs 0
	IsEnabled        *bool   `json:"is_enabled"`                                // Pointer handles null/omitted vs false
	Status           *string `json:"status"`                                    // Optional override for initial status
}

type CheckHandler struct {
	CheckRepo repository.CheckRepository
}

// NewCheckHandler creates a new CheckHandler with necessary dependencies.
// >>> Add this constructor function <<<
func NewCheckHandler(cr repository.CheckRepository) *CheckHandler {
	return &CheckHandler{CheckRepo: cr}
}

func (h *CheckHandler) CreateCheck(c *gin.Context) {
	var req CreateCheckRequest // <<< Bind to the request struct >>>

	// 1. Bind JSON and Validate using binding tags
	if err := c.ShouldBindJSON(&req); err != nil {
		// Gin's binding validation provides helpful error messages
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 2. Get User ID (from auth middleware context)
	userIDtmp, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		log.Println("ERROR: UserID not found in context for protected route /api/v1/checks")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication context error",
		})
		return
	}
	// The middleware returns int, your model might use int64, so cast if needed
	userID := int64(userIDtmp)

	// 3. Map data from Request struct to DB Model struct
	newCheck := models.Check{
		UserID:           userID,
		UUID:             uuid.NewString(), // Generate UUID here
		Name:             req.Name,         // Directly assign required fields
		ExpectedInterval: req.ExpectedInterval,
		// Set defaults for optional/nullable fields first
		IsEnabled: true,  // Default to enabled
		Status:    "new", // Default to new status
	}

	// Populate optional fields from request if they were provided
	if req.Description != nil {
		newCheck.Description = sql.NullString{String: *req.Description, Valid: true}
	} // Otherwise, Description remains sql.NullString{Valid: false} (NULL)

	if req.GracePeriod != nil {
		newCheck.GracePeriod = *req.GracePeriod
	} // Otherwise, GracePeriod remains 0

	if req.IsEnabled != nil {
		newCheck.IsEnabled = *req.IsEnabled // Override default if provided
	}
	if req.Status != nil {
		// TODO: Add validation here if you only allow specific status values initially
		newCheck.Status = *req.Status // Override default if provided
	}

	// 4. Call Repository Create method with the populated models.Check
	ctx := c.Request.Context()
	err := h.CheckRepo.Create(ctx, &newCheck) // Pass pointer to the models.Check struct

	if err != nil {
		if errors.Is(err, repository.ErrCheckNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Related resource not found"})
		} else if strings.Contains(err.Error(), "already exists") { // Basic duplicate check
			c.JSON(http.StatusConflict, gin.H{"error": "Check with this UUID might already exist"})
		} else {
			log.Printf("ERROR: CreateCheck handler failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create check"})
		}
		return
	}

	// 5. Return Success Response (using the populated models.Check struct)
	c.JSON(http.StatusCreated, newCheck)
}

func (h *CheckHandler) GetChecks(c *gin.Context) {
	// 1. Get User ID (from auth middleware context)
	userIDtmp, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		log.Println("ERROR: UserID not found in context for protected route /api/v1/checks")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication context error",
		})
		return
	}
	userID := int64(userIDtmp)
	log.Printf("INFO: GetChecks request received for user ID: %d", userID)

	// 2. Call Repository List method
	ctx := c.Request.Context()
	checks, err := h.CheckRepo.ListByUserID(ctx, userID)

	// 3. Handle Repository Errors
	if err != nil {
		// It's NOT an error if the user simply has no checks.
		// sql.ErrNoRows is often not returned for list queries that find nothing,
		// they usually return an empty slice and nil error.
		// However, if your repository method specifically returns ErrCheckNotFound or similar, handle it.
		if errors.Is(err, repository.ErrCheckNotFound) {
			log.Printf("INFO: No checks found for user ID: %d", userID)
			c.JSON(http.StatusOK, []models.Check{})
			return
		}

		// Handle other potential database errors
		log.Printf("ERROR: GetChecks handler repository call failed for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve checks",
		})
		return
	}

	// Handle the case where the query runs fine but finds no rows (returns empty slice, nil error)
	if checks == nil {
		// Ensure we always return a JSON array, even if empty, not null
		checks = []models.Check{}
	}

	// 4. Return Success Response
	log.Printf("INFO: Successfully retrieved %d checks for user ID: %d", len(checks), userID)
	c.JSON(http.StatusOK, checks)
}

func (h *CheckHandler) UpdateCheck(c *gin.Context) {

}

func (h *CheckHandler) DeleteCheck(c *gin.Context) {

}

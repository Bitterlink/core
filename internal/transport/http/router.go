package httptransport

import (
	"bitterlink/core/internal/middleware"
	"bitterlink/core/internal/repository"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all the application routes
func RegisterRoutes(
	router *gin.Engine,
	pingHandler *PingHandler,
	checkHandler *CheckHandler,
	dbPool *sql.DB,
	repo repository.CheckRepository,
) {
	// --- Public Routes ---
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to PING",
			"status":  "running",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"server_time": time.Now().UTC().Format(time.RFC3339Nano),
		})
	})

	// --- API v1 Routes ---
	apiV1 := router.Group("/api/v1")

	apiV1.Use(middleware.APIKeyAuthMiddleware(dbPool))
	{
		// Check management endpoints
		apiV1.GET("/ping/:uuid", pingHandler.HandlePing)
		apiV1.POST("/checks", checkHandler.CreateCheck)
		apiV1.GET("/checks", checkHandler.GetChecks)
	}
}

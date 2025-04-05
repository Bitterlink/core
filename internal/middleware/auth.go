package middleware

import (
	"bitterlink/core/internal/agency"
	//"context"
	"database/sql"
	"errors" // Import errors package
	"log"
	"net/http"
	"strings"

	// "nova/ping/internal/models" // Assuming you have a models package for User/APIKey structs

	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID" // Key to store/retrieve user ID from Gin context

// APIKeyAuthMiddleware creates a Gin middleware handler for API key authentication.
// It requires a database connection pool to validate keys.
func APIKeyAuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("WRN: Authorization header missing")
			// Optional: Add WWW-Authenticate header for standard compliance
			c.Header("WWW-Authenticate", `Bearer realm="api"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}
		// 2. Check if it's a Bearer token
		const bearerPrefix = "Bearer"
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			log.Printf("WARN: Invalid Authorization header format (missing '%s' prefix)", bearerPrefix)
			c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_token", error_description="Authorization header format must be Bearer {token}"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}

		// 3. Extract the token (API key) itself
		apiKey := strings.TrimPrefix(authHeader, bearerPrefix)
		if apiKey == "" {
			log.Println("WARN: Authorization header present but token is empty")
			c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_token", error_description="Bearer token is empty"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token is empty",
			})
			return
		}

		// 2. Validate the key against the database
		// IMPORTANT SECURITY NOTE: In production, you should HASH API keys in the database
		// and compare hashes, not plaintext keys. This example uses plaintext for simplicity.
		var userID int
		var isActive bool

		query := "SELECT user_id, is_active FROM api_keys WHERE key_value = ? LIMIT 1"
		err := db.QueryRowContext(c.Request.Context(), query, strings.TrimSpace(apiKey)).Scan(&userID, &isActive)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Key not found
				log.Printf("WARN: Invalid API key presented via Bearer token: %s...", apiKey[:agency.Min(len(apiKey), 10)]) // Log prefix only
				c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_token", error_description="Invalid API key"`)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid API key",
				})
				return
			}
			// Other database error
			log.Printf("ERROR: Database error during API key validation: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Could not validate API key",
			})
			return
		}

		// 3. Check if the key is active
		if !isActive {
			log.Printf("WARN: Inactive API key presented for user %d", userID)
			c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_token", error_description="API key is inactive"`)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "API key is inactive",
			})
			return
		}

		// 4. Store User ID in context for downsteam handlers
		c.Set(UserIDKey, userID)
		log.Printf("INFO: API key validated successfully for user %d", userID)
		// 5. Call the next handler in the chain
		c.Next()
	}
}

// GetUserIDFromContext retrieves the user ID stored in the Gin context by the middleware.
// Returns the user ID and true if found, otherwise 0 and false.
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userIDVal, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	userID, ok := userIDVal.(int)
	if !ok {
		log.Printf("ERROR: UserID in context is not an int: %T", userIDVal)
		return 0, false
	}
	return userID, true
}

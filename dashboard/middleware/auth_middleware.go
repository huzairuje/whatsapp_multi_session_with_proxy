package middleware

import (
	"net/http"
	"strings"
	"whatsapp_multi_session/dashboard/auth" // Path to the auth package

	"github.com/gin-gonic/gin"
)

// JWTMiddleware creates a Gin middleware for JWT authentication.
// It takes the JWT secret key as a parameter.
func JWTMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secretKey == "" {
			// This is a server configuration error, should ideally be caught at startup
			// Log.Error("JWT secret key is not configured for middleware")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Authentication system not configured"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") { // Using EqualFold for case-insensitive "Bearer"
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenString := parts[1]

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token string cannot be empty"})
			return
		}

		claims, err := auth.ValidateJWT(tokenString, secretKey)
		if err != nil {
			// Log the error for debugging on the server side if needed, e.g., log.Printf("JWT validation error: %v", err)
			// The specific error from ValidateJWT (e.g., expired, malformed) is already wrapped.
			// We return a generic error to the client.
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid, expired, or malformed token"})
			return
		}

		// Set user information in the context for downstream handlers
		// The types for AdminUserID and PhoneNumber in JWTClaims are uint and string, respectively.
		// Gin's c.Set stores interface{}, so these will be retrieved with type assertion later.
		c.Set("admin_user_id", claims.AdminUserID)
		c.Set("phone_number", claims.PhoneNumber)

		c.Next()
	}
}

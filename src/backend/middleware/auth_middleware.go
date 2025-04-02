package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Eduardo-Barreto/web-ponderada/backend/auth"
	"github.com/Eduardo-Barreto/web-ponderada/backend/utils"
)

// AuthMiddleware checks for a valid JWT in the Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Check if the header format is "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.SendError(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, err.Error()) // Use error message from ValidateToken
			c.Abort()
			return
		}

		// Add claims (like user ID) to the context for handlers to use
		c.Set("userID", claims.UserID)
        c.Set("userEmail", claims.Email) // Can be useful

		c.Next() // Proceed to the next handler
	}
}

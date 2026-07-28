package middleware

import (
	"net/http"
	"strings"

	"ark-server-commander/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Prefer Authorization header (standard for REST requests)
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token format"})
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else if q := c.Query("token"); q != "" {
			// Fallback: token via query param (required for browser WebSocket clients,
			// which cannot set custom headers on the initial WS upgrade request)
			tokenString = q
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		// ParseAccessToken, not ParseToken: a refresh token is valid for 30 days and
		// must not be replayable as a bearer token on request-authentication paths
		claims, err := utils.ParseAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

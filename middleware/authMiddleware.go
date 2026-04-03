package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 1. RequireAuth: Checks if the user is logged in (Valid Token)
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header sent by your React frontend
		authHeader := c.GetHeader("Authorization")

		// It should look like "Bearer eyJhbGciOi..."
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: No token provided"})
			c.Abort() // Stops the request from continuing
			return
		}

		// Extract just the token string
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Make sure the signing method is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract the data (Claims) we packed into the token during login
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Save the user's ID and Role into the Gin context so the next function can use it
			c.Set("userID", claims["sub"])
			c.Set("userRole", claims["role"])

			c.Next() // Pass the request to the actual route handler
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: Invalid token payload"})
			c.Abort()
			return
		}
	}
}

// 2. RequireAdmin: Strict RBAC check (Must be placed AFTER RequireAuth)
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Grab the role that RequireAuth just saved into the context
		role, exists := c.Get("userRole")

		if !exists || role != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden: Admin access required"})
			c.Abort()
			return
		}

		c.Next() // Let the Admin through!
	}
}

package routes

import (
	"net/http"

	"bm-pharmacy-api/controllers" // IMPORT ADDED HERE
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.Engine) {
	// Create a group for all admin routes
	adminGroup := router.Group("/api/v1/admin")

	// Apply BOTH middlewares to this group.
	// 1st: Check if they are logged in. 2nd: Check if they are an Admin.
	adminGroup.Use(middleware.RequireAuth(), middleware.RequireAdmin())
	{
		// Because of the middleware above, no one can even reach this code
		// unless they possess a valid JWT token with the role "ADMIN"

		adminGroup.GET("/dashboard", func(c *gin.Context) {
			// We can safely assume they are an admin here!
			c.JSON(http.StatusOK, gin.H{
				"message": "Welcome to the Admin Dashboard",
				"stats":   "Top secret sales data here",
			})
		})

		// --- USER MANAGEMENT ROUTES ---
		adminGroup.GET("/users", controllers.GetAllUsers)
		adminGroup.PUT("/users/:id/role", controllers.UpdateUserRole)

		// You would add your analytics, inventory, and review moderation routes here...
	}
}

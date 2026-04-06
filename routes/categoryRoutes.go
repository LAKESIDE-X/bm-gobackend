package routes

import (
	"bm-pharmacy-api/controllers"

	"github.com/gin-gonic/gin"
)

func CategoryRoutes(router *gin.Engine) {
	categoryGroup := router.Group("/api/v1/categories")
	{
		// Public Route
		categoryGroup.GET("", controllers.GetCategories)

		// Admin Routes (You can wrap these in Admin Middleware later!)
		categoryGroup.POST("", controllers.CreateCategory)
		categoryGroup.DELETE("/:id", controllers.DeleteCategory)
	}
}

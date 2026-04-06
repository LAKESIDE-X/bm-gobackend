package routes

import (
	"bm-pharmacy-api/controllers"

	"github.com/gin-gonic/gin"
)

func BrandRoutes(router *gin.Engine) {
	brandGroup := router.Group("/api/v1/brands")
	{
		// Public Route
		brandGroup.GET("", controllers.GetBrands)

		// Admin Routes
		brandGroup.POST("", controllers.CreateBrand)
		brandGroup.DELETE("/:id", controllers.DeleteBrand)
	}
}

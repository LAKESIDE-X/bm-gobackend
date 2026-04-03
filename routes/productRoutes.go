package routes

import (
	"bm-pharmacy-api/controllers"
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func ProductRoutes(router *gin.Engine) {
	// ==========================================
	// PUBLIC ROUTES (Anyone can view products)
	// ==========================================
	publicGroup := router.Group("/api/v1/products")
	{
		publicGroup.GET("", controllers.GetProducts)        // Get all active products
		publicGroup.GET("/:id", controllers.GetProductByID) // Get one product by ID
	}

	// ==========================================
	// PROTECTED ADMIN ROUTES (Only Admins can edit)
	// ==========================================
	// We reuse the /api/v1/admin prefix
	adminGroup := router.Group("/api/v1/admin/products")

	// Apply the bouncers!
	adminGroup.Use(middleware.RequireAuth(), middleware.RequireAdmin())
	{
		adminGroup.POST("", controllers.CreateProduct)       // Create a new product
		adminGroup.PATCH("/:id", controllers.UpdateProduct)  // Update an existing product
		adminGroup.DELETE("/:id", controllers.DeleteProduct) // Delete a product
	}

	productGroup := router.Group("/api/v1/admin/products")
	{
		// Make sure this matches the API path your React app is calling!
		productGroup.POST("/", controllers.CreateProduct)
	}
}

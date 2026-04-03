// routes/orderRoutes.go
package routes

import (
	"bm-pharmacy-api/controllers"
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func OrderRoutes(router *gin.Engine) {
	// Group all customer order routes
	orderGroup := router.Group("/api/v1/orders")
	orderGroup.Use(middleware.RequireAuth())
	{
		orderGroup.POST("/checkout", controllers.Checkout)
		orderGroup.GET("", controllers.GetUserOrders)
		// You can add the /:id and /:id/cancel routes here later!
	}
}

// routes/cartRoutes.go
package routes

import (
	"bm-pharmacy-api/controllers"
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func CartRoutes(router *gin.Engine) {
	// Group all cart routes and protect them with RequireAuth
	cartGroup := router.Group("/api/v1/cart")
	cartGroup.Use(middleware.RequireAuth())
	{
		cartGroup.GET("", controllers.GetCart)
		cartGroup.POST("/add", controllers.AddToCart)
		cartGroup.PATCH("/item/:productId", controllers.UpdateCartItem)
		cartGroup.DELETE("/item/:productId", controllers.RemoveFromCart)
		cartGroup.DELETE("", controllers.ClearCart)
	}
}

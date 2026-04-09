package routes

import (
	"bm-pharmacy-api/controllers" // Update this if your module name is different!
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func WishlistRoutes(router *gin.Engine) {
	wishlistGroup := router.Group("/api/v1/wishlist")

	// ALL wishlist routes require the user to be logged in!
	wishlistGroup.Use(middleware.RequireAuth())
	{
		wishlistGroup.GET("", controllers.GetWishlist)
		wishlistGroup.POST("", controllers.AddToWishlist)
		wishlistGroup.DELETE("/:id", controllers.RemoveFromWishlist)
		wishlistGroup.GET("/check/:id", controllers.CheckWishlistStatus)
	}
}

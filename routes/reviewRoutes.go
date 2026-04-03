package routes

import (
	"bm-pharmacy-api/controllers"
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func ReviewRoutes(router *gin.Engine) {

	// PUBLIC ROUTE: Anyone can read product reviews
	// Note: Changed :productId to :id to match ProductRoutes
	router.GET("/api/v1/products/:id/reviews", controllers.GetProductReviews)

	// PROTECTED ROUTES: Only logged-in users can write reviews or check eligibility
	protectedGroup := router.Group("/api/v1/products")
	protectedGroup.Use(middleware.RequireAuth())
	{
		// Note: Changed :productId to :id to match ProductRoutes
		protectedGroup.GET("/:id/reviews/can-review", controllers.CheckCanReview)
		protectedGroup.POST("/:id/reviews", controllers.CreateReview)
	}
}

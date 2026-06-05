package routes

import (
	"bm-pharmacy-api/controllers"
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func ReviewRoutes(router *gin.Engine) {
	// --- PUBLIC ROUTE ---
	// Anyone can read reviews on a product page (matching frontend: /products/:productId/reviews)
	router.GET("/api/v1/products/:id/reviews", controllers.GetProductReviews)

	// --- PROTECTED USER ROUTES ---
	// Must be logged in to leave a review (matching frontend: /products/:productId/reviews)
	userGroup := router.Group("/api/v1/products/:id/reviews")
	userGroup.Use(middleware.RequireAuth())
	{
		userGroup.POST("", controllers.AddReview)
	}

	// Check if user can review a product
	router.GET("/api/v1/products/:id/reviews/can-review", middleware.RequireAuth(), controllers.CanReviewProduct)

	// Get user's own reviews
	router.GET("/api/v1/reviews/me", middleware.RequireAuth(), controllers.GetMyReviews)

	// Update and delete user's own reviews
	router.PATCH("/api/v1/reviews/:id", middleware.RequireAuth(), controllers.UpdateReview)
	router.DELETE("/api/v1/reviews/:id", middleware.RequireAuth(), controllers.DeleteReview)

	// --- PROTECTED ADMIN ROUTES ---
	// Must be logged in AND be an Admin to view all or delete
	adminGroup := router.Group("/api/v1/admin/reviews")
	adminGroup.Use(middleware.RequireAuth(), middleware.RequireAdmin())
	{
		adminGroup.GET("", controllers.AdminGetAllReviews)
		adminGroup.DELETE("/:id", controllers.AdminDeleteReview)
	}
}

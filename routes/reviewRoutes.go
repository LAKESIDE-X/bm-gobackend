package routes

import (
	"bm-pharmacy-api/controllers"
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func ReviewRoutes(router *gin.Engine) {
	// --- PUBLIC ROUTE ---
	// Anyone can read reviews on a product page
	router.GET("/api/v1/reviews/product/:id", controllers.GetProductReviews)

	// --- PROTECTED USER ROUTES ---
	// Must be logged in to leave a review
	userGroup := router.Group("/api/v1/reviews")
	userGroup.Use(middleware.RequireAuth())
	{
		userGroup.POST("", controllers.AddReview)
	}

	// --- PROTECTED ADMIN ROUTES ---
	// Must be logged in AND be an Admin to view all or delete
	adminGroup := router.Group("/api/v1/admin/reviews")
	adminGroup.Use(middleware.RequireAuth(), middleware.RequireAdmin())
	{
		adminGroup.GET("", controllers.AdminGetAllReviews)
		adminGroup.DELETE("/:id", controllers.AdminDeleteReview)
	}
}

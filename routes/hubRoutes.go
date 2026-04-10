package routes

import (
	"bm-pharmacy-api/controllers"
	"bm-pharmacy-api/middleware"

	"github.com/gin-gonic/gin"
)

func HubRoutes(router *gin.Engine) {
	// --- PUBLIC ROUTES (Anyone can read articles and subscribe) ---
	router.POST("/api/v1/newsletter/subscribe", controllers.SubscribeNewsletter)
	router.GET("/api/v1/articles", controllers.GetArticles)
	router.GET("/api/v1/articles/:id", controllers.GetArticleByID)

	// --- ADMIN ROUTES (Only Admins can publish or delete) ---
	adminGroup := router.Group("/api/v1/admin/hub")
	adminGroup.Use(middleware.RequireAuth(), middleware.RequireAdmin())
	{
		adminGroup.GET("/subscribers", controllers.GetSubscribers)
		adminGroup.POST("/articles", controllers.CreateArticle)
		adminGroup.DELETE("/articles/:id", controllers.DeleteArticle)
	}
}

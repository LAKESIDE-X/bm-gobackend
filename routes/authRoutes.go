package routes

import (
	"bm-pharmacy-api/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	// Grouping all auth routes under /api/v1/auth
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", controllers.Register)
		authGroup.POST("/login", controllers.Login)
		authGroup.POST("/forgot-password", controllers.ForgotPassword)
		authGroup.POST("/reset-password", controllers.ResetPassword)
		authGroup.POST("/send-registration-otp", controllers.SendRegistrationOTP)
	}
}

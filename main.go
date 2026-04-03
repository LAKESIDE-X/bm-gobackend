package main

import (
	"log"
	"os"
	"time"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Connect to the Database
	database.ConnectDB()

	// 3. Initialize the Gin router
	router := gin.Default()
	router.Static("/uploads", "./uploads")

	// --- NEW OFFICIAL CORS CONFIGURATION ---
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Explicitly trust your React app!
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// ---------------------------------------

	// 4. Health Check Route
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "BM Pharmacy API is running smoothly! 🚀",
		})
	})

	// 5. Register All Routes
	routes.AuthRoutes(router)
	routes.ProductRoutes(router)
	routes.ReviewRoutes(router)
	routes.CartRoutes(router)
	routes.OrderRoutes(router)
	routes.AdminRoutes(router) // Assuming you created this earlier

	// 6. Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback for localhost
	}
	router.Run(":" + port)
}

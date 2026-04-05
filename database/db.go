package database

import (
	"fmt"
	"log"
	"os"

	"bm-pharmacy-api/models" // This links to your models folder

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// 1. First, try to grab the full URL (This is what Render uses)
	dsn := os.Getenv("DB_URL")

	// 2. If DB_URL is empty, we must be testing locally, so build it manually
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Lagos",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
	}

	// 3. Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database! \n", err)
	}

	log.Println("Database connection successfully opened!")

	// 4. Run Migrations
	log.Println("Running Database Migrations...")
	err = db.AutoMigrate(&models.User{}, &models.Product{}, &models.CartItem{}, &models.Order{}, &models.OrderItem{}, &models.Review{}, &models.OTP{})
	if err != nil {
		log.Fatal("Failed to migrate database: \n", err)
	}
	log.Println("Database Migrations Completed!")

	DB = db
}

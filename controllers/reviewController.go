package controllers

import (
	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// We create a specific helper here to avoid naming conflicts with your other files
func getReviewUserID(c *gin.Context) uint {
	userVal, exists := c.Get("userID")
	if !exists {
		return 0
	}
	switch v := userVal.(type) {
	case models.User:
		return v.ID
	case *models.User:
		return v.ID
	case float64:
		return uint(v)
	case uint:
		return v
	case int:
		return uint(v)
	default:
		return 0
	}
}

// 1. GET: Fetch all reviews for a specific product (Public)
func GetProductReviews(c *gin.Context) {
	productID := c.Param("id")
	var reviews []models.Review

	// Preload "User" so we can display the reviewer's first and last name!
	if err := database.DB.Preload("User").Where("product_id = ?", productID).Order("created_at desc").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reviews})
}

// 2. POST: Add a new review (Requires Login)
func AddReview(c *gin.Context) {
	userID := getReviewUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user session"})
		return
	}

	var input struct {
		ProductID uint   `json:"productId" binding:"required"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Comment   string `json:"comment" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Optional: Stop users from spamming multiple reviews on the same product
	var existing models.Review
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, input.ProductID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already reviewed this product."})
		return
	}

	review := models.Review{
		UserID:    userID,
		ProductID: input.ProductID,
		Rating:    input.Rating,
		Comment:   input.Comment,
	}

	database.DB.Create(&review)

	// Fetch it back immediately with the User data attached so React can display it instantly
	database.DB.Preload("User").First(&review, review.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Review added successfully", "data": review})
}

// 3. GET: Fetch ALL reviews across the store (Admin Only)
func AdminGetAllReviews(c *gin.Context) {
	var reviews []models.Review
	// Preload User (who wrote it) and Product (what they reviewed)
	database.DB.Preload("User").Preload("Product").Order("created_at desc").Find(&reviews)

	c.JSON(http.StatusOK, gin.H{"data": reviews})
}

// 4. DELETE: Remove an inappropriate review (Admin Only)
func AdminDeleteReview(c *gin.Context) {
	reviewID := c.Param("id")
	database.DB.Delete(&models.Review{}, reviewID)
	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}

package controllers

import (
	"bm-pharmacy-api/database" // Updated to match your folder!
	"bm-pharmacy-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET: Fetch the user's wishlist
func GetWishlist(c *gin.Context) {
	// Get the logged-in user's ID from the Auth Middleware
	userID, _ := c.Get("user")

	var wishlist []models.Wishlist
	// Preload("Product") automatically fetches the product images, names, and prices!
	database.DB.Preload("Product").Where("user_id = ?", userID).Find(&wishlist)

	c.JSON(http.StatusOK, gin.H{"data": wishlist})
}

// POST: Add a product to the wishlist
func AddToWishlist(c *gin.Context) {
	userID, _ := c.Get("user")

	var input struct {
		ProductID uint `json:"productId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if it already exists so we don't add duplicates
	var existing models.Wishlist
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, input.ProductID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Already in wishlist"})
		return
	}

	wishlist := models.Wishlist{
		UserID:    userID.(uint),
		ProductID: input.ProductID,
	}

	database.DB.Create(&wishlist)
	c.JSON(http.StatusOK, gin.H{"message": "Added to wishlist", "data": wishlist})
}

// DELETE: Remove a product
func RemoveFromWishlist(c *gin.Context) {
	userID, _ := c.Get("user")
	productID := c.Param("id")

	database.DB.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.Wishlist{})
	c.JSON(http.StatusOK, gin.H{"message": "Removed from wishlist"})
}

// GET: Quick check if a product is in the wishlist (for the Heart icon to light up)
func CheckWishlistStatus(c *gin.Context) {
	userID, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, false)
		return
	}
	productID := c.Param("id")

	var count int64
	database.DB.Model(&models.Wishlist{}).Where("user_id = ? AND product_id = ?", userID, productID).Count(&count)

	// Returns true if the count is greater than 0
	c.JSON(http.StatusOK, count > 0)
}

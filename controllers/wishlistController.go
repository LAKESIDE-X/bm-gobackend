package controllers

import (
	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// --- BULLETPROOF HELPER FUNCTION ---
// Safely extracts the User ID no matter how the Auth Middleware saved it
func getSafeUserID(c *gin.Context) uint {
	userVal, exists := c.Get("user")
	if !exists {
		return 0
	}

	switch v := userVal.(type) {
	case models.User:
		return v.ID // Middleware saved the whole struct
	case *models.User:
		return v.ID // Middleware saved a pointer to the struct
	case float64:
		return uint(v) // Middleware saved a JWT parsed number
	case uint:
		return v // Middleware saved a perfect uint
	case int:
		return uint(v) // Middleware saved a standard int
	default:
		return 0
	}
}

// GET: Fetch the user's wishlist
func GetWishlist(c *gin.Context) {
	userID := getSafeUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user session"})
		return
	}

	var wishlist []models.Wishlist
	database.DB.Preload("Product").Where("user_id = ?", userID).Find(&wishlist)

	c.JSON(http.StatusOK, gin.H{"data": wishlist})
}

// POST: Add a product to the wishlist
func AddToWishlist(c *gin.Context) {
	userID := getSafeUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user session"})
		return
	}

	var input struct {
		ProductID uint `json:"productId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Wishlist
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, input.ProductID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Already in wishlist"})
		return
	}

	wishlist := models.Wishlist{
		UserID:    userID,
		ProductID: input.ProductID,
	}

	database.DB.Create(&wishlist)
	c.JSON(http.StatusOK, gin.H{"message": "Added to wishlist", "data": wishlist})
}

// DELETE: Remove a product
func RemoveFromWishlist(c *gin.Context) {
	userID := getSafeUserID(c)
	productID := c.Param("id")

	database.DB.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.Wishlist{})
	c.JSON(http.StatusOK, gin.H{"message": "Removed from wishlist"})
}

// GET: Quick check if a product is in the wishlist
func CheckWishlistStatus(c *gin.Context) {
	userID := getSafeUserID(c)
	if userID == 0 {
		c.JSON(http.StatusOK, false)
		return
	}

	productID := c.Param("id")

	var count int64
	database.DB.Model(&models.Wishlist{}).Where("user_id = ? AND product_id = ?", userID, productID).Count(&count)

	c.JSON(http.StatusOK, count > 0)
}

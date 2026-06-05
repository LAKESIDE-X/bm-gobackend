package controllers

import (
	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// --- BULLETPROOF HELPER FUNCTION ---
// Safely extracts the User ID no matter how the Auth Middleware saved it
func getSafeUserID(c *gin.Context) uint {
	// THE FIX IS HERE: Changed "user" to "userID" to match your middleware!
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

// GET: Fetch the user's wishlist
func GetWishlist(c *gin.Context) {
	userID := getSafeUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user session"})
		return
	}

	var wishlist []models.Wishlist
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	// Count total wishlist items
	var total int64
	database.DB.Model(&models.Wishlist{}).Where("user_id = ?", userID).Count(&total)

	// Get paginated wishlist
	offset := (parseInt(page) - 1) * parseInt(limit)
	err := database.DB.Preload("Product").Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(parseInt(limit)).
		Offset(offset).
		Find(&wishlist).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	totalPages := (total + int64(parseInt(limit)) - 1) / int64(parseInt(limit))

	c.JSON(http.StatusOK, gin.H{
		"items": wishlist,
		"pagination": gin.H{
			"total":      total,
			"page":       parseInt(page),
			"limit":      parseInt(limit),
			"totalPages": totalPages,
		},
	})
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

// GET: Get wishlist count
func GetWishlistCount(c *gin.Context) {
	userID := getSafeUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user session"})
		return
	}

	var count int64
	database.DB.Model(&models.Wishlist{}).Where("user_id = ?", userID).Count(&count)

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// POST: Move item from wishlist to cart
func MoveToCart(c *gin.Context) {
	userID := getSafeUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user session"})
		return
	}

	productID := c.Param("id")

	// Check if product exists in wishlist
	var wishlistItem models.Wishlist
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&wishlistItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in wishlist"})
		return
	}

	// Check if product already exists in cart
	var existingCartItem models.CartItem
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&existingCartItem).Error; err == nil {
		// Update quantity
		database.DB.Model(&existingCartItem).Update("quantity", existingCartItem.Quantity+1)
	} else {
		// Add to cart
		cartItem := models.CartItem{
			UserID:    userID,
			ProductID: wishlistItem.ProductID,
			Quantity:  1,
		}
		database.DB.Create(&cartItem)
	}

	// Remove from wishlist
	database.DB.Delete(&wishlistItem)

	// Get updated cart
	var cart []models.CartItem
	database.DB.Preload("Product").Where("user_id = ?", userID).Find(&cart)

	c.JSON(http.StatusOK, gin.H{
		"message": "Item moved to cart",
		"cart":    cart,
	})
}

// Helper function to parse string to int
func parseInt(s string) int {
	var i int
	_, _ = fmt.Sscanf(s, "%d", &i)
	return i
}

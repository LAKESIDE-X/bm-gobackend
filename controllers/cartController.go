// controllers/cartController.go
package controllers

import (
	"net/http"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"

	"github.com/gin-gonic/gin"
)

type AddCartInput struct {
	ProductID uint `json:"productId" binding:"required"`
	Quantity  int  `json:"quantity"`
}

type UpdateCartInput struct {
	Quantity int `json:"quantity" binding:"required"`
}

// Helper to safely get the User ID from the JWT token context
func getUserID(c *gin.Context) uint {
	// JWT stores numbers as float64, so we assert it and convert to uint
	userIDFloat, _ := c.Get("userID")
	return uint(userIDFloat.(float64))
}

// 1. GET CART
func GetCart(c *gin.Context) {
	userID := getUserID(c)
	var cartItems []models.CartItem

	// Preload("Product") is GORM's magic trick. It fetches the Product table data
	// and injects it into the CartItem response so the frontend gets names and prices!
	if err := database.DB.Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": cartItems})
}

// 2. ADD TO CART
func AddToCart(c *gin.Context) {
	userID := getUserID(c)
	var input AddCartInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	// Set a default quantity if none is provided
	if input.Quantity <= 0 {
		input.Quantity = 1
	}

	var cartItem models.CartItem

	// Check if this user already has this exact product in their cart
	err := database.DB.Where("user_id = ? AND product_id = ?", userID, input.ProductID).First(&cartItem).Error

	if err == nil {
		// Item exists! Just increase the quantity.
		cartItem.Quantity += input.Quantity
		database.DB.Save(&cartItem)
	} else {
		// Item doesn't exist, create a brand new cart item
		cartItem = models.CartItem{
			UserID:    userID,
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		}
		database.DB.Create(&cartItem)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item added to cart", "cartItem": cartItem})
}

// 3. UPDATE QUANTITY
func UpdateCartItem(c *gin.Context) {
	userID := getUserID(c)
	productID := c.Param("productId") // The frontend passes the Product ID in the URL
	var input UpdateCartInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	var cartItem models.CartItem
	// Find the specific item for THIS user
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&cartItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Item not found in cart"})
		return
	}

	// Update and save
	cartItem.Quantity = input.Quantity
	database.DB.Save(&cartItem)

	c.JSON(http.StatusOK, gin.H{"message": "Quantity updated", "cartItem": cartItem})
}

// 4. REMOVE ITEM
func RemoveFromCart(c *gin.Context) {
	userID := getUserID(c)
	productID := c.Param("productId")

	// Delete where User ID and Product ID match
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.CartItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to remove item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
}

// 5. CLEAR ENTIRE CART
func ClearCart(c *gin.Context) {
	userID := getUserID(c)

	// Delete ALL items that belong to this user
	if err := database.DB.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
}

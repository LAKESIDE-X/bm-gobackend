// controllers/cartController.go
package controllers

import (
	"net/http"
	"strings"
	"time"

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

// Helper to build frontend-compatible cart response
func buildCartResponse(userID uint, cartItems []models.CartItem) gin.H {
	itemCount := 0
	totalAmount := 0.0
	items := make([]gin.H, 0, len(cartItems))

	for _, item := range cartItems {
		itemCount += item.Quantity
		subtotal := item.Product.Price * float64(item.Quantity)
		totalAmount += subtotal

		items = append(items, gin.H{
			"id":        item.ID,
			"productId": item.ProductID,
			"quantity":  item.Quantity,
			"product": gin.H{
				"id":            item.Product.ID,
				"name":          item.Product.Name,
				"slug":          slugify(item.Product.Name),
				"price":         item.Product.Price,
				"thumbnailUrl":  item.Product.ImageURL,
				"stockQuantity": item.Product.Stock,
			},
			"subtotal": subtotal,
		})
	}

	return gin.H{
		"id":          userID,
		"userId":      userID,
		"items":       items,
		"itemCount":   itemCount,
		"totalAmount": totalAmount,
		"createdAt":   time.Now(),
		"updatedAt":   time.Now(),
	}
}

func slugify(name string) string {
	// Simple slug generation - replace spaces with hyphens and lowercase
	result := ""
	for _, ch := range name {
		if ch == ' ' {
			result += "-"
		} else if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			result += string(ch)
		} else if ch == '-' || ch == '_' {
			result += string(ch)
		}
	}
	if result == "" {
		return "product"
	}
	return strings.ToLower(result)
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

	c.JSON(http.StatusOK, gin.H{"cart": buildCartResponse(userID, cartItems)})
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

	// Get updated cart
	var cartItems []models.CartItem
	database.DB.Preload("Product").Where("user_id = ?", userID).Find(&cartItems)

	c.JSON(http.StatusOK, gin.H{"message": "Item added to cart", "cart": buildCartResponse(userID, cartItems)})
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

	// Get updated cart
	var cartItems []models.CartItem
	database.DB.Preload("Product").Where("user_id = ?", userID).Find(&cartItems)

	c.JSON(http.StatusOK, gin.H{"message": "Quantity updated", "cart": buildCartResponse(userID, cartItems)})
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

	// Get updated cart
	var cartItems []models.CartItem
	database.DB.Preload("Product").Where("user_id = ?", userID).Find(&cartItems)

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart", "cart": buildCartResponse(userID, cartItems)})
}

// 5. CLEAR ENTIRE CART
func ClearCart(c *gin.Context) {
	userID := getUserID(c)

	// Delete ALL items that belong to this user
	if err := database.DB.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully", "cart": buildCartResponse(userID, nil)})
}

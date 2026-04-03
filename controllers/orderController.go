// controllers/orderController.go
package controllers

import (
	"net/http"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CheckoutInput struct {
	ShippingAddress string `json:"shippingAddress" binding:"required"`
	PaymentMethod   string `json:"paymentMethod" binding:"required"`
}

// 1. CHECKOUT (Cart -> Order)
func Checkout(c *gin.Context) {
	userID := getUserID(c) // Reusing the helper from our cartController!
	var input CheckoutInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid checkout data"})
		return
	}

	// Start a Database Transaction
	// If any step returns an error, the entire transaction "Rolls Back" (cancels).
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch the user's cart items with the Product data included
		var cartItems []models.CartItem
		if err := tx.Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error; err != nil {
			return err
		}

		if len(cartItems) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Your cart is empty"})
			return nil // Returning nil here because we manually handled the error response
		}

		var totalAmount float64
		var orderItems []models.OrderItem

		// 2. Loop through cart, calculate total, and check inventory stock!
		for _, item := range cartItems {
			// CRITICAL: Check if we actually have enough stock
			if item.Product.Stock < item.Quantity {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Not enough stock for " + item.Product.Name})
				return nil
			}

			totalAmount += item.Product.Price * float64(item.Quantity)

			// Prep the OrderItem
			orderItems = append(orderItems, models.OrderItem{
				ProductID:       item.ProductID,
				Quantity:        item.Quantity,
				PriceAtPurchase: item.Product.Price,
			})

			// 3. Deduct the stock from the Product table!
			newStock := item.Product.Stock - item.Quantity
			if err := tx.Model(&item.Product).Update("stock", newStock).Error; err != nil {
				return err
			}
		}

		// 4. Create the overarching Order record
		order := models.Order{
			UserID:          userID,
			TotalAmount:     totalAmount,
			ShippingAddress: input.ShippingAddress,
			PaymentMethod:   input.PaymentMethod,
			Status:          "PENDING", // Status starts as pending until paid/shipped
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// 5. Link the OrderItems to this new Order and save them
		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return err
		}

		// 6. Clear the user's Cart since they just bought everything!
		if err := tx.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		// Send success response
		c.JSON(http.StatusCreated, gin.H{
			"message": "Order placed successfully",
			"orderId": order.ID,
		})

		return nil // Returning nil tells GORM to "Commit" (save) the transaction
	})

	// If the transaction failed, send a generic error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process checkout", "error": err.Error()})
	}
}

// 2. GET USER ORDERS (Order History)
func GetUserOrders(c *gin.Context) {
	userID := getUserID(c)
	var orders []models.Order

	// Preload the OrderItems AND the Product details inside those OrderItems!
	if err := database.DB.Preload("OrderItems.Product").Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

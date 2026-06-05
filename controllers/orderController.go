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

// Helper to transform backend Order to frontend-compatible format
func buildOrderResponse(order models.Order) gin.H {
	itemCount := 0
	items := make([]gin.H, 0, len(order.OrderItems))
	for _, item := range order.OrderItems {
		itemCount += item.Quantity
		items = append(items, gin.H{
			"productId": item.ProductID,
			"name":      item.Product.Name,
			"quantity":  item.Quantity,
			"unitPrice": item.PriceAtPurchase,
			"subtotal":  item.PriceAtPurchase * float64(item.Quantity),
		})
	}

	paymentStatus := "PENDING"
	if order.Status == "PAID" || order.Status == "PROCESSING" || order.Status == "SHIPPED" || order.Status == "DELIVERED" {
		paymentStatus = "SUCCESS"
	} else if order.Status == "CANCELLED" {
		paymentStatus = "FAILED"
	}

	// Try to parse shipping address as JSON object, fallback to string wrap
	var shippingAddress interface{} = order.ShippingAddress

	return gin.H{
		"id":              order.ID,
		"status":          order.Status,
		"totalAmount":     order.TotalAmount,
		"itemCount":       itemCount,
		"items":           items,
		"paymentStatus":   paymentStatus,
		"shippingAddress": shippingAddress,
		"createdAt":       order.CreatedAt,
		"updatedAt":       order.UpdatedAt,
	}
}

// 2. GET USER ORDERS (Order History)
func GetUserOrders(c *gin.Context) {
	userID := getUserID(c)
	var orders []models.Order
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")
	status := c.Query("status")

	// Count total orders
	query := database.DB.Model(&models.Order{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)

	// Get paginated orders
	query = database.DB.Preload("OrderItems.Product").Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	offset := (parseInt(page) - 1) * parseInt(limit)
	if err := query.Order("created_at desc").Limit(parseInt(limit)).Offset(offset).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch orders"})
		return
	}

	totalPages := (total + int64(parseInt(limit)) - 1) / int64(parseInt(limit))

	// Transform orders to frontend format
	transformedOrders := make([]gin.H, 0, len(orders))
	for _, order := range orders {
		transformedOrders = append(transformedOrders, buildOrderResponse(order))
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": transformedOrders,
		"pagination": gin.H{
			"total":      total,
			"page":       parseInt(page),
			"limit":      parseInt(limit),
			"totalPages": totalPages,
		},
	})
}

// 3. GET ORDER BY ID
func GetOrderById(c *gin.Context) {
	userID := getUserID(c)
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.Preload("OrderItems.Product").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, buildOrderResponse(order))
}

// 4. CANCEL ORDER
func CancelOrder(c *gin.Context) {
	userID := getUserID(c)
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}

	// Only allow cancellation if order is pending
	if order.Status != "PENDING" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot cancel an order that is not pending"})
		return
	}

	// Update order status
	if err := database.DB.Model(&order).Update("status", "CANCELLED").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to cancel order"})
		return
	}

	// Restore stock for each item
	for _, item := range order.OrderItems {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err == nil {
			database.DB.Model(&product).Update("stock", product.Stock+item.Quantity)
		}
	}

	// Refresh order data with items
	database.DB.Preload("OrderItems.Product").First(&order, order.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully", "data": buildOrderResponse(order)})
}

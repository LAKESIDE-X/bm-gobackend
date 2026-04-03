package controllers

import (
	"net/http"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"

	"github.com/gin-gonic/gin"
)

type ReviewInput struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"required"`
}

// 1. CHECK IF USER CAN REVIEW (The Gatekeeper)
func CheckCanReview(c *gin.Context) {
	userID := getUserID(c)

	// Changed from "productId" to "id" to match the route update
	productID := c.Param("id")

	var count int64

	// THE LOGIC: Look inside the "orders" table. Join it with the "order_items" table.
	// Does this User have an Order with status 'DELIVERED' that includes this ProductID?
	database.DB.Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Where("orders.user_id = ? AND orders.status = ? AND order_items.product_id = ?", userID, "DELIVERED", productID).
		Count(&count)

	// If count > 0, they bought it and received it.
	if count > 0 {
		c.JSON(http.StatusOK, true)
	} else {
		c.JSON(http.StatusOK, false)
	}
}

// 2. CREATE A REVIEW
func CreateReview(c *gin.Context) {
	userID := getUserID(c)

	// Changed from "productId" to "id" to match the route update
	productID := c.Param("id")

	var input ReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid review data (Rating must be 1-5)"})
		return
	}

	// First, check if they already reviewed this product to prevent spam
	var existingReview models.Review
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&existingReview).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "You have already reviewed this product."})
		return
	}

	// Double-check they actually bought it (Server-side security)
	var count int64
	database.DB.Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Where("orders.user_id = ? AND orders.status = ? AND order_items.product_id = ?", userID, "DELIVERED", productID).
		Count(&count)

	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"message": "You can only review products you have purchased and received."})
		return
	}

	// All checks passed! Save the review.
	var product models.Product
	database.DB.First(&product, productID)

	review := models.Review{
		ProductID: product.ID,
		UserID:    userID,
		Rating:    input.Rating,
		Comment:   input.Comment,
	}

	if err := database.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to submit review"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Review submitted successfully!", "data": review})
}

// 3. GET ALL REVIEWS FOR A PRODUCT (Public)
func GetProductReviews(c *gin.Context) {
	// Changed from "productId" to "id" to match the route update
	productID := c.Param("id")
	var reviews []models.Review

	// Preload the User so the frontend can display the reviewer's First Name!
	if err := database.DB.Preload("User").Where("product_id = ?", productID).Order("created_at desc").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reviews})
}

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
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	// Count total reviews
	var total int64
	database.DB.Model(&models.Review{}).Where("product_id = ?", productID).Count(&total)

	// Get paginated reviews
	offset := (parseInt(page) - 1) * parseInt(limit)
	if err := database.DB.Preload("User").Where("product_id = ?", productID).
		Order("created_at desc").
		Limit(parseInt(limit)).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	// Calculate review stats
	averageRating := 0.0
	if total > 0 {
		sum := 0
		for _, r := range reviews {
			sum += r.Rating
		}
		averageRating = float64(sum) / float64(total)
	}

	// Get rating breakdown for all reviews on this product (not just paginated)
	var allReviews []models.Review
	database.DB.Where("product_id = ?", productID).Find(&allReviews)
	ratingBreakdown := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	for _, r := range allReviews {
		if r.Rating >= 1 && r.Rating <= 5 {
			ratingBreakdown[r.Rating]++
		}
	}

	totalPages := (total + int64(parseInt(limit)) - 1) / int64(parseInt(limit))

	c.JSON(http.StatusOK, gin.H{
		"reviews":         reviews,
		"averageRating":   averageRating,
		"totalReviews":    len(allReviews),
		"ratingBreakdown": ratingBreakdown,
		"pagination": gin.H{
			"total":      total,
			"page":       parseInt(page),
			"limit":      parseInt(limit),
			"totalPages": totalPages,
		},
	})
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

// 5. GET: Check if user can review a product
func CanReviewProduct(c *gin.Context) {
	userID := getReviewUserID(c)
	productID := c.Param("id")

	// Check if user has purchased this product
	var orderCount int64
	database.DB.Table("order_items").
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Where("orders.user_id = ? AND order_items.product_id = ?", userID, productID).
		Count(&orderCount)

	// Check if user has already reviewed this product
	var reviewCount int64
	database.DB.Model(&models.Review{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&reviewCount)

	canReview := orderCount > 0 && reviewCount == 0
	reason := ""
	if !canReview {
		if orderCount == 0 {
			reason = "You must purchase this product before reviewing it"
		} else {
			reason = "You have already reviewed this product"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"canReview": canReview,
		"reason":    reason,
	})
}

// 6. GET: Get user's own reviews
func GetMyReviews(c *gin.Context) {
	userID := getReviewUserID(c)
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	var reviews []models.Review
	var total int64

	// Count total reviews
	database.DB.Model(&models.Review{}).Where("user_id = ?", userID).Count(&total)

	// Get paginated reviews
	offset := (parseInt(page) - 1) * parseInt(limit)
	if err := database.DB.Preload("Product").Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(parseInt(limit)).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	totalPages := (total + int64(parseInt(limit)) - 1) / int64(parseInt(limit))

	c.JSON(http.StatusOK, gin.H{
		"reviews": reviews,
		"pagination": gin.H{
			"total":      total,
			"page":       parseInt(page),
			"limit":      parseInt(limit),
			"totalPages": totalPages,
		},
	})
}

// 7. PATCH: Update a review
func UpdateReview(c *gin.Context) {
	userID := getReviewUserID(c)
	reviewID := c.Param("id")

	var review models.Review
	if err := database.DB.First(&review, reviewID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	// Check if user owns this review
	if review.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own reviews"})
		return
	}

	var input struct {
		Rating  *int    `json:"rating"`
		Comment *string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only provided fields
	updates := make(map[string]interface{})
	if input.Rating != nil {
		updates["rating"] = *input.Rating
	}
	if input.Comment != nil {
		updates["comment"] = *input.Comment
	}

	if err := database.DB.Model(&review).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	// Fetch updated review with relations
	database.DB.Preload("User").Preload("Product").First(&review, reviewID)

	c.JSON(http.StatusOK, gin.H{"message": "Review updated successfully", "data": review})
}

// 8. DELETE: Delete a review (user can delete their own)
func DeleteReview(c *gin.Context) {
	userID := getReviewUserID(c)
	reviewID := c.Param("id")

	var review models.Review
	if err := database.DB.First(&review, reviewID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	// Check if user owns this review
	if review.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own reviews"})
		return
	}

	database.DB.Delete(&review)

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}

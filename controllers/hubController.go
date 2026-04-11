package controllers

import (
	"net/http"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"

	"github.com/gin-gonic/gin"
)

// ==========================================
// NEWSLETTER SUBSCRIPTIONS
// ==========================================

// POST: Add a new email to the newsletter list (Public)
func SubscribeNewsletter(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide a valid email address."})
		return
	}

	subscriber := models.Subscriber{Email: input.Email}

	// Create the subscriber. If duplicate, GORM will return an error,
	// but we return success to the user to prevent email harvesting/spam.
	if err := database.DB.Create(&subscriber).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Thanks for subscribing!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thanks for subscribing!"})
}

// GET: Fetch all subscribers (Admin Only)
func GetSubscribers(c *gin.Context) {
	var subscribers []models.Subscriber

	if err := database.DB.Order("created_at desc").Find(&subscribers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscribers})
}

// ==========================================
// WELLNESS HUB ARTICLES
// ==========================================

// GET: Fetch all articles (Public)
func GetArticles(c *gin.Context) {
	var articles []models.Article

	// Pre-sorting by newest first so the Hub looks fresh
	if err := database.DB.Order("created_at desc").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": articles})
}

// GET: Fetch a single article by ID (Public)
func GetArticleByID(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := database.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

// POST: Upload a new article (Admin Only)
// CORRECTED: Now accepts a permanent Cloudinary URL string from React
func CreateArticle(c *gin.Context) {
	title := c.PostForm("title")
	category := c.PostForm("category")
	content := c.PostForm("content")

	// React uploads the image to Cloudinary first and sends the resulting URL here
	imageURL := c.PostForm("imageUrl")

	// Basic Validation
	if title == "" || category == "" || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, category, and content are required"})
		return
	}

	article := models.Article{
		Title:    title,
		Category: category,
		Content:  content,
		ImageURL: imageURL, // This is now a permanent Cloudinary link
	}

	if err := database.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save article to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Article published successfully",
		"data":    article,
	})
}

// DELETE: Remove an article (Admin Only)
func DeleteArticle(c *gin.Context) {
	id := c.Param("id")

	if err := database.DB.Delete(&models.Article{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
}

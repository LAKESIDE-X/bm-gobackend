package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

	// If the email already exists, just return success anyway (good for security/spam prevention)
	if err := database.DB.Create(&subscriber).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Thanks for subscribing!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thanks for subscribing!"})
}

// GET: Fetch all subscribers (Admin Only)
func GetSubscribers(c *gin.Context) {
	var subscribers []models.Subscriber
	database.DB.Order("created_at desc").Find(&subscribers)
	c.JSON(http.StatusOK, gin.H{"data": subscribers})
}

// ==========================================
// WELLNESS HUB ARTICLES
// ==========================================

// GET: Fetch all articles (Public)
func GetArticles(c *gin.Context) {
	var articles []models.Article
	database.DB.Order("created_at desc").Find(&articles)
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

// POST: Upload a new article with a cover image (Admin Only)
func CreateArticle(c *gin.Context) {
	title := c.PostForm("title")
	category := c.PostForm("category")
	content := c.PostForm("content")

	if title == "" || category == "" || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, category, and content are required"})
		return
	}

	var imagePath string

	// Handle the cover image upload
	file, err := c.FormFile("image")
	if err == nil {
		// Make sure the directory exists
		os.MkdirAll("./uploads/articles", os.ModePerm)

		// Create a unique filename
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		fullPath := "./uploads/articles/" + filename

		if err := c.SaveUploadedFile(file, fullPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
		// Save the relative path for React to use
		imagePath = "/uploads/articles/" + filename
	}

	article := models.Article{
		Title:    title,
		Category: category,
		Content:  content,
		ImageURL: imagePath,
	}

	database.DB.Create(&article)
	c.JSON(http.StatusOK, gin.H{"message": "Article published successfully", "data": article})
}

// DELETE: Remove an article (Admin Only)
func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.Article{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
}

package controllers

import (
	"net/http"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"

	"github.com/gin-gonic/gin"
)

// 1. CREATE CATEGORY (Admin Only)
func CreateCategory(c *gin.Context) {
	var category models.Category

	// Since we aren't uploading images here, we can just use simple JSON binding!
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data format"})
		return
	}

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Category created successfully",
		"data":    category,
	})
}

// 2. GET ALL CATEGORIES (Public)
func GetCategories(c *gin.Context) {
	var categories []models.Category

	if err := database.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}

// 3. DELETE CATEGORY (Admin Only)
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Category not found"})
		return
	}

	database.DB.Delete(&category)

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}

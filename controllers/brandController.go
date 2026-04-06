package controllers

import (
	"net/http"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"

	"github.com/gin-gonic/gin"
)

// 1. CREATE BRAND (Admin Only)
func CreateBrand(c *gin.Context) {
	var brand models.Brand

	if err := c.ShouldBindJSON(&brand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data format"})
		return
	}

	if err := database.DB.Create(&brand).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create brand"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Brand created successfully",
		"data":    brand,
	})
}

// 2. GET ALL BRANDS (Public)
func GetBrands(c *gin.Context) {
	var brands []models.Brand

	if err := database.DB.Find(&brands).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch brands"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": brands,
	})
}

// 3. DELETE BRAND (Admin Only)
func DeleteBrand(c *gin.Context) {
	id := c.Param("id")
	var brand models.Brand

	if err := database.DB.First(&brand, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Brand not found"})
		return
	}

	database.DB.Delete(&brand)

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand deleted successfully",
	})
}

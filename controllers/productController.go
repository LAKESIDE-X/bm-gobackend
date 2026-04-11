package controllers

import (
	"net/http"
	"strconv"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"

	"github.com/gin-gonic/gin"
)

// 1. CREATE PRODUCT (Admin Only - Now accepts Cloudinary URL from Frontend)
func CreateProduct(c *gin.Context) {
	// Parse the text fields from the FormData
	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	stockStr := c.PostForm("stock")
	categoryIDStr := c.PostForm("categoryId")
	brandIDStr := c.PostForm("brandId")

	// NEW: We now just take the ImageURL as a string from the frontend
	// React will handle the Cloudinary upload and pass the link here
	imageURL := c.PostForm("imageUrl")

	// Convert strings to numbers
	price, _ := strconv.ParseFloat(priceStr, 64)
	stock, _ := strconv.Atoi(stockStr)
	categoryID, _ := strconv.Atoi(categoryIDStr)
	brandID, _ := strconv.Atoi(brandIDStr)

	// Assemble the Product model
	product := models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		CategoryID:  uint(categoryID),
		BrandID:     uint(brandID),
		ImageURL:    imageURL, // This is now a permanent Cloudinary link!
		IsActive:    true,
	}

	// Save to Database
	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save product to database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": product,
	})
}

// 2. GET ALL PRODUCTS (Public)
func GetProducts(c *gin.Context) {
	var products []models.Product

	if err := database.DB.Preload("Category").Preload("Brand").Where("is_active = ?", true).Order("created_at desc").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

// 3. GET SINGLE PRODUCT (Public)
func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := database.DB.Preload("Category").Preload("Brand").First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": product,
	})
}

// 4. UPDATE PRODUCT (Admin Only)
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	// 1. Check if the product exists first
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	// 2. Use a Map to catch the incoming data.
	// This stops the "Invalid update data" error because maps don't care about strict naming!
	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON format"})
		return
	}

	// 3. Update the record using GORM's .Updates()
	// This will only change the fields you actually sent from React
	if err := database.DB.Model(&product).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update product in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": product,
	})
}

// 5. DELETE PRODUCT (Admin Only)
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	database.DB.Delete(&product)

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}

package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"

	"github.com/gin-gonic/gin"
)

// 1. CREATE PRODUCT (Admin Only - Handles Image Uploads)
func CreateProduct(c *gin.Context) {
	// Parse the text fields from the FormData
	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	stockStr := c.PostForm("stock")
	categorySlug := c.PostForm("categorySlug")
	brandSlug := c.PostForm("brandSlug")

	// Convert Price and Stock to numbers
	price, _ := strconv.ParseFloat(priceStr, 64)
	stock, _ := strconv.Atoi(stockStr)

	// Handle the Image Upload
	var imageURL string
	file, err := c.FormFile("image")

	if err == nil {
		// Create the uploads folder if it doesn't exist
		uploadDir := "uploads/products"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		// Create a unique filename using the current timestamp
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		filePath := filepath.Join(uploadDir, filename)

		// Save the physical file to our server
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save image"})
			return
		}

		// The URL path we will save in the database so React can fetch it
		imageURL = "/" + filepath.ToSlash(filePath)
	} else {
		// If no image was uploaded, assign an empty string (or a placeholder path later)
		imageURL = ""
	}

	// Assemble the Product model
	product := models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Category:    categorySlug,
		Brand:       brandSlug,
		ImageURL:    imageURL,
		IsActive:    true, // Default to active
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

	// Only fetch products where IsActive is true.
	// We use GORM's Find method to get an array of records.
	if err := database.DB.Where("is_active = ?", true).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch products"})
		return
	}

	// Your React frontend expects this exact structure!
	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

// 3. GET SINGLE PRODUCT (Public)
func GetProductByID(c *gin.Context) {
	// Grab the ID from the URL (e.g., /api/v1/products/5)
	id := c.Param("id")
	var product models.Product

	if err := database.DB.First(&product, id).Error; err != nil {
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

	// First, check if the product actually exists
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	// Bind the incoming JSON to update the fields
	// Note: If you want to update images later, this function will need to be updated to use FormData too!
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid update data"})
		return
	}

	// Save the changes back to the database
	database.DB.Save(&product)

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

	// Because we added gorm.DeletedAt to our Product model earlier,
	// this doesn't actually erase it from the hard drive. It does a "Soft Delete",
	// meaning it just hides it, keeping your past order histories perfectly intact!
	database.DB.Delete(&product)

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}

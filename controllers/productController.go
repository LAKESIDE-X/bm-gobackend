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

	// UPGRADE: We now expect the IDs of the Category and Brand from the frontend dropdowns
	categoryIDStr := c.PostForm("categoryId")
	brandIDStr := c.PostForm("brandId")

	// Convert strings to numbers
	price, _ := strconv.ParseFloat(priceStr, 64)
	stock, _ := strconv.Atoi(stockStr)
	categoryID, _ := strconv.Atoi(categoryIDStr)
	brandID, _ := strconv.Atoi(brandIDStr)

	// Handle the Image Upload
	var imageURL string
	file, err := c.FormFile("image")

	if err == nil {
		uploadDir := "uploads/products"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		filePath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save image"})
			return
		}

		imageURL = "/" + filepath.ToSlash(filePath)
	} else {
		imageURL = ""
	}

	// Assemble the Product model using the new Foreign Keys
	product := models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		CategoryID:  uint(categoryID), // Assigning the ID to the relation
		BrandID:     uint(brandID),    // Assigning the ID to the relation
		ImageURL:    imageURL,
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

	// UPGRADE: Added .Preload("Category") and .Preload("Brand") to fetch the related data!
	if err := database.DB.Preload("Category").Preload("Brand").Where("is_active = ?", true).Find(&products).Error; err != nil {
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

	// UPGRADE: Preload here too!
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

	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid update data"})
		return
	}

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

	database.DB.Delete(&product)

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}

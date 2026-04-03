package controllers

import (
	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 1. GET ALL USERS (Admin Only)
func GetAllUsers(c *gin.Context) {
	var users []models.User

	// We only select safe columns (DO NOT SEND PASSWORDS TO THE FRONTEND!)
	if err := database.DB.Select("id, first_name, last_name, email, phone, role, created_at").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

// Struct to catch the new role
type UpdateRoleInput struct {
	Role string `json:"role" binding:"required"`
}

// 2. PROMOTE / DEMOTE USER (Admin Only)
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var input UpdateRoleInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	// Security check: Only allow valid roles
	if input.Role != "admin" && input.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Role must be 'admin' or 'user'"})
		return
	}

	// Update the database
	if err := database.DB.Model(&models.User{}).Where("id = ?", id).Update("role", input.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully!"})
}

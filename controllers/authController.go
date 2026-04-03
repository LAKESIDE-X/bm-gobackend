package controllers

import (
	"net/http"
	"time"

	"bm-pharmacy-api/database"
	"bm-pharmacy-api/models"
	"bm-pharmacy-api/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Struct to catch incoming Login data
type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Struct to safely catch incoming Registration data
type RegisterInput struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone"`
	Password  string `json:"password" binding:"required,min=6"`
	OTP       string `json:"otp" binding:"required"`
}

type SendOTPInput struct {
	Email string `json:"email" binding:"required,email"`
}

// -----------------------------------------
// 1. SEND REGISTRATION OTP
// -----------------------------------------
func SendRegistrationOTP(c *gin.Context) {
	var input SendOTPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Valid email is required"})
		return
	}

	// Make sure the email isn't already registered
	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already in use"})
		return
	}

	// Generate OTP and save to DB
	otpCode := utils.GenerateOTP()
	otpRecord := models.OTP{
		Email:     input.Email,
		Code:      otpCode,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	database.DB.Create(&otpRecord)

	// Send Email (Synchronous to catch errors!)
	emailBody := "Welcome to BM Pharmacy!\n\nYour registration verification code is: " + otpCode + "\n\nThis code expires in 15 minutes."
	err := utils.SendEmail(input.Email, "Registration Verification Code", emailBody)

	if err != nil {
		// If email fails, delete the OTP from the database and alert the frontend
		database.DB.Delete(&otpRecord)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent to your email!"})
}

// -----------------------------------------
// 2. REGISTER NEW USER
// -----------------------------------------
func Register(c *gin.Context) {
	var input RegisterInput

	// Read the JSON using our safe struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input data", "error": err.Error()})
		return
	}

	// CHECK OTP BEFORE CONTINUING
	var otpRecord models.OTP
	if err := database.DB.Where("email = ? AND code = ? AND expires_at > ?", input.Email, input.OTP, time.Now()).First(&otpRecord).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Incorrect or expired verification code."})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already in use"})
		return
	}

	// Hash the actual password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password"})
		return
	}

	// Assemble the final User model
	user := models.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Password:  string(hashedPassword),
	}

	// Save to Database
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user"})
		return
	}

	// Delete the used OTP so it cannot be reused
	database.DB.Delete(&otpRecord)

	// Send success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"user":    user,
	})
}

// -----------------------------------------
// 3. LOGIN USER
// -----------------------------------------
func Login(c *gin.Context) {
	var input LoginInput

	// Read the JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input data"})
		return
	}

	// Find the user by email
	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	// Check if password matches the hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	// Generate JWT Token
	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
		return
	}

	// Send token to frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordInput struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// -----------------------------------------
// 4. FORGOT PASSWORD (Send OTP)
// -----------------------------------------
func ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Valid email is required"})
		return
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// For security, we don't tell the user if the email exists or not
		c.JSON(http.StatusOK, gin.H{"message": "If that email exists, an OTP has been sent."})
		return
	}

	// Generate OTP and save to DB
	otpCode := utils.GenerateOTP()
	otpRecord := models.OTP{
		Email:     user.Email,
		Code:      otpCode,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	database.DB.Create(&otpRecord)

	// Send Email (Synchronous to catch errors!)
	emailBody := "Your BM Pharmacy password reset code is: " + otpCode + "\n\nThis code expires in 15 minutes."
	err := utils.SendEmail(user.Email, "Password Reset Code", emailBody)

	if err != nil {
		// Clean up DB if email fails
		database.DB.Delete(&otpRecord)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, an OTP has been sent."})
}

// -----------------------------------------
// 5. RESET PASSWORD (Verify OTP & Update DB)
// -----------------------------------------
func ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input data"})
		return
	}

	// Check if OTP is valid and not expired
	var otpRecord models.OTP
	if err := database.DB.Where("email = ? AND code = ? AND expires_at > ?", input.Email, input.OTP, time.Now()).First(&otpRecord).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid or expired OTP code"})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to secure new password"})
		return
	}

	// Update the user's password in the DB
	database.DB.Model(&models.User{}).Where("email = ?", input.Email).Update("password", string(hashedPassword))

	// Delete the used OTP so it can't be reused
	database.DB.Delete(&otpRecord)

	c.JSON(http.StatusOK, gin.H{"message": "Password successfully reset!"})
}

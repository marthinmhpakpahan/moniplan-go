package controllers

import (
	"net/http"
	"strings"

	"ashborn.id/moniplan/database"
	"ashborn.id/moniplan/middlewares"
	"ashborn.id/moniplan/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRequest struktur untuk validasi input register
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// LoginRequest struktur untuk validasi input login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse struktur untuk response authentication
type AuthResponse struct {
	Message string            `json:"message"`
	User    models.PublicUser `json:"user"`
	Token   string            `json:"token"`
}

// Register handler untuk membuat user baru
func Register(c *gin.Context) {
	var req RegisterRequest

	// Validasi dan bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	// Normalize email ke lowercase
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// Cek apakah email sudah terdaftar
	var existingUser models.User
	result := database.DB.Where("email = ?", req.Email).First(&existingUser)

	if result.Error == nil {
		// User dengan email ini sudah ada
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Email already registered",
			"message": "An account with this email already exists",
		})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		// Error database lainnya
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"message": "Failed to check existing user",
		})
		return
	}

	// Buat user baru
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password, // Akan di-hash oleh BeforeCreate hook
	}

	// Simpan ke database
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Registration failed",
			"message": "Failed to create user account",
		})
		return
	}

	// Generate JWT token untuk auto-login setelah register
	token, err := middlewares.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Account created but failed to generate authentication token",
		})
		return
	}

	// Success response
	c.JSON(http.StatusCreated, AuthResponse{
		Message: "Registration successful",
		User:    user.ToPublicUser(),
		Token:   token,
	})
}

// Login handler untuk autentikasi user
func Login(c *gin.Context) {
	var req LoginRequest

	// Validasi dan bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Cari user berdasarkan email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User tidak ditemukan
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Invalid email or password",
			})
		} else {
			// Database error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"message": "Failed to fetch user data",
			})
		}
		return
	}

	// Verifikasi password
	if err := user.CheckPassword(req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"message": "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, err := middlewares.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Login successful but failed to generate authentication token",
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, AuthResponse{
		Message: "Login successful",
		User:    user.ToPublicUser(),
		Token:   token,
	})
}

// GetProfile handler untuk mendapatkan data user yang sedang login
func GetProfile(c *gin.Context) {
	// Ambil user ID dari context (di-set oleh auth middleware)
	userID, exists := middlewares.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	// Fetch user data dari database
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "User account no longer exists",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"message": "Failed to fetch user profile",
			})
		}
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile fetched successfully",
		"user":    user.ToPublicUser(),
	})
}

// RefreshToken handler untuk refresh JWT token
func RefreshToken(c *gin.Context) {
	// Ambil user info dari context
	userID, _ := middlewares.GetUserIDFromContext(c)
	userEmail, _ := c.Get("userEmail")
	userName, _ := c.Get("userName")

	// Generate token baru
	token, err := middlewares.GenerateToken(
		userID,
		userEmail.(string),
		userName.(string),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Failed to refresh authentication token",
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"token":   token,
	})
}

// HealthCheck handler untuk cek status API
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "API is running",
	})
}

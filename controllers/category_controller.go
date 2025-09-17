package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"ashborn.id/moniplan/database"
	"ashborn.id/moniplan/middlewares"
	"ashborn.id/moniplan/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryRequest struct {
	UserID     uint   `json:"user_id" binding:"required"`
	CategoryID uint   `json:"category_id"`
	Name       string `json:"name" binding:"required,max=100"`
	Month      uint   `json:"month" binding:"required"`
	Year       uint   `json:"year" binding:"required"`
	Amount     uint   `json:"amount" binding:"required"`
}

type CategoryDefaultResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type CategoryIndexResponse struct {
	Error   bool              `json:"error"`
	Message string            `json:"message"`
	Data    []models.Category `json:"data"`
}

type CategoryFetchResponse struct {
	Error        bool            `json:"error"`
	Message      string          `json:"message"`
	DataCategory models.Category `json:"data_category"`
	DataBudget   models.Budget   `json:"data_budget"`
}

func IndexCategory(c *gin.Context) {
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
	var categories []models.Category
	if err := database.DB.Where("user_id = ?", userID).Order("name ASC").Find(&categories).Error; err != nil {
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
	c.JSON(http.StatusCreated, CategoryIndexResponse{
		Error:   false,
		Message: "Category & Budget creation successful",
		Data:    categories,
	})
}

func CreateCategory(c *gin.Context) {
	var req CategoryRequest

	// Validasi dan bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	req.Name = strings.ToLower(req.Name)

	now := time.Now()

	// Cek apakah email sudah terdaftar
	var existingCategory models.Category
	result := database.DB.Where("name = ?", req.Name).Where("user_id = ?", req.UserID).First(&existingCategory)

	if result.Error == nil {
		req.CategoryID = existingCategory.ID
	} else {
		// Buat user baru
		newCategory := models.Category{
			UserID:    req.UserID,
			Name:      req.Name,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := database.DB.Create(&newCategory).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Category creation failed",
				"message": err,
			})
			return
		}
	}

	newBudget := models.Budget{
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		Month:      req.Month,
		Year:       req.Year,
		Amount:     req.Amount,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := database.DB.Create(&newBudget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Budget creation failed",
			"message": err,
		})
	}

	// Success response
	c.JSON(http.StatusCreated, CategoryDefaultResponse{
		Error:   false,
		Message: "Category & Budget creation successful",
	})
}

func GetCategoryByID(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	var category models.Category
	var budget models.Budget
	year, month, _ := time.Now().Date()

	userID, exists := middlewares.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid Category ID!",
		})
		return
	} else {
		if errCategory := database.DB.First(&category, categoryID).Error; errCategory != nil {
			if errCategory == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Category not found",
					"message": "Category no longer exists",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Database error",
					"message": "Failed to fetch category",
				})
			}
			return
		}

		if errBudget := database.DB.Where("category_id = ? AND user_id = ? AND year = ? AND month = ?", categoryID, userID, year, month).Last(&budget).Error; errBudget != nil {
			if errBudget == gorm.ErrRecordNotFound {
				if errBudget = database.DB.Where("category_id = ? AND user_id = ?", categoryID, userID).Last(&budget).Error; errBudget != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Database error",
						"message": "Failed to fetch budget",
					})
					return
				}
			}
		}
	}

	// Success response
	c.JSON(http.StatusCreated, CategoryFetchResponse{
		Error:        false,
		Message:      "Category & Budget creation successful",
		DataCategory: category,
		DataBudget:   budget,
	})
}

func UpdateCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Required Param",
			"message": err.Error(),
		})
	}

	var req CategoryRequest

	// Validasi dan bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	req.Name = strings.ToLower(req.Name)

	now := time.Now()

	var existingCategory models.Category
	result := database.DB.First(&existingCategory, categoryID)

	if result.Error == nil {
		req.CategoryID = existingCategory.ID
	} else {
		newCategory := models.Category{
			UserID:    req.UserID,
			Name:      req.Name,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := database.DB.Create(&newCategory).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Category creation failed",
				"message": err,
			})
			return
		}
	}

	newBudget := models.Budget{
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		Month:      req.Month,
		Year:       req.Year,
		Amount:     req.Amount,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := database.DB.Create(&newBudget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Budget creation failed",
			"message": err,
		})
	}

	// Success response
	c.JSON(http.StatusCreated, CategoryDefaultResponse{
		Error:   false,
		Message: "Category & Budget successfully updated",
	})
}

func DeleteCategoryByID(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid Category ID!",
		})
		return
	}

	if err := database.DB.Delete(&models.Category{}, categoryID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed",
			"message": "Unable to delete category!",
		})
		return
	}

	// Success response
	c.JSON(http.StatusCreated, CategoryDefaultResponse{
		Error:   false,
		Message: "Category deletion successful",
	})
}

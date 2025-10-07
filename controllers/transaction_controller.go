package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ashborn.id/moniplan/database"
	"ashborn.id/moniplan/middlewares"
	"ashborn.id/moniplan/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionRequest struct {
	TransactionID   uint   `json:"transaction_id"`
	CategoryID      uint   `json:"category_id"`
	Amount          uint   `json:"amount" binding:"required"`
	Type            string `json:"type" binding:"required,max=100"`
	Remarks         string `json:"remarks" binding:"required"`
	TransactionDate string `json:"transaction_date"`
}

type UpdateTransactionRequest struct {
	TransactionID   uint   `json:"transaction_id"`
	CategoryID      uint   `json:"category_id"`
	Amount          uint   `json:"amount"`
	Type            string `json:"type" binding:"max=100"`
	Remarks         string `json:"remarks" binding:"max=255"`
	TransactionDate string `json:"transaction_date"`
}

type TransactionDefaultResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type TransactionIndexResponse struct {
	Error   bool                               `json:"error"`
	Message string                             `json:"message"`
	Data    []models.TransactionCategoryBudget `json:"data"`
}

type TransactionFetchResponse struct {
	Error   bool               `json:"error"`
	Message string             `json:"message"`
	Data    models.Transaction `json:"data"`
}

func IndexTransaction(c *gin.Context) {
	// Ambil user ID dari context (di-set oleh auth middleware)
	userID, exists := middlewares.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	now := time.Now() // current local time
	currentYear, currentMonth, _ := now.Date()
	current_year := currentYear
	current_month := int(currentMonth)

	param_month := c.DefaultQuery("month", strconv.Itoa(current_month))
	param_year := c.DefaultQuery("year", strconv.Itoa(current_year))

	// Fetch user data dari database
	var transactions []models.TransactionCategoryBudget
	err := database.DB.
		Table("transactions t").
		Select("t.id, t.user_id, t.category_id, t.amount, t.type, t.remarks, DATE_FORMAT(t.transaction_date, '%W, %d %M %Y %H:%i') AS transaction_date, t.created_at, t.updated_at, c.name as category_name").
		Joins("JOIN categories c ON c.id = t.category_id").
		Where("t.user_id = ? AND YEAR(t.transaction_date) = ? AND MONTH(t.transaction_date) = ?", userID, param_year, param_month).
		Order("t.id DESC").
		Scan(&transactions).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Transaction not found",
				"message": "Transaction no longer exists",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"message": "Failed to fetch transaction",
			})
		}
		return
	}

	// Success response
	c.JSON(http.StatusCreated, TransactionIndexResponse{
		Error:   false,
		Message: "Data loaded!",
		Data:    transactions,
	})
}

func GetTransactionByID(c *gin.Context) {
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	var transaction models.Transaction

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid Category ID!",
		})
		return
	} else {
		if errTransaction := database.DB.First(&transaction, transactionID).Error; errTransaction != nil {
			if errTransaction == gorm.ErrRecordNotFound {
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
	}

	// Success response
	c.JSON(http.StatusCreated, TransactionFetchResponse{
		Error:   false,
		Message: "Category & Budget creation successful",
		Data:    transaction,
	})
}

func CreateTransaction(c *gin.Context) {
	var req TransactionRequest

	// Ambil user ID dari context (di-set oleh auth middleware)
	userID, exists := middlewares.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	// Validasi dan bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	now := time.Now()

	_datetime, err := time.Parse(req.TransactionDate, req.TransactionDate)
	if err != nil {
		_datetime = now
	}

	newTransaction := models.Transaction{
		UserID:          userID,
		CategoryID:      req.CategoryID,
		Amount:          req.Amount,
		Type:            req.Type,
		Remarks:         req.Remarks,
		TransactionDate: _datetime,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := database.DB.Create(&newTransaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Transaction creation failed",
			"message": err,
		})
		return
	}

	// Success response
	c.JSON(http.StatusCreated, CategoryDefaultResponse{
		Error:   false,
		Message: "Transaction creation successful",
	})
}

func UpdateTransaction(c *gin.Context) {
	var req UpdateTransactionRequest

	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Required Param",
			"message": err.Error(),
		})
	}

	// Validasi dan bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	fmt.Println("Transaction ID: ", transactionID)

	now := time.Now()

	// Cek apakah email sudah terdaftar
	var existingTransaction models.Transaction
	result := database.DB.First(&existingTransaction, transactionID)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Transaction not found!",
			"message": err,
		})
		return
	}

	existingTransaction.UpdatedAt = now

	if req.CategoryID > 0 {
		existingTransaction.CategoryID = req.CategoryID
	}

	if req.Amount > 0 {
		existingTransaction.Amount = req.Amount
	}

	if req.TransactionDate != "" {
		layout := "2006-01-02 15:04:05"
		transaction_date, _ := time.Parse(layout, req.TransactionDate)
		existingTransaction.TransactionDate = transaction_date
	}

	if req.Remarks != "" {
		existingTransaction.Remarks = req.Remarks
	}

	fmt.Println("Amount: ", req.Amount)

	if err := database.DB.Updates(existingTransaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error in updating transaction!",
			"message": err,
		})
		return
	}

	// Success response
	c.JSON(http.StatusCreated, CategoryDefaultResponse{
		Error:   false,
		Message: "Transaction successfully updated",
	})
}

func DeleteTransactionByID(c *gin.Context) {
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid Category ID!",
		})
		return
	}

	if err := database.DB.Delete(&models.Transaction{}, transactionID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed",
			"message": "Unable to delete transaction!",
		})
		return
	}

	// Success response
	c.JSON(http.StatusCreated, TransactionDefaultResponse{
		Error:   false,
		Message: "Transaction deletion successful",
	})
}

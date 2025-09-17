package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func EmptyController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error":   "No error",
		"message": "Still In Progress",
	})
}

package routes

import (
	"ashborn.id/moniplan/controllers"
	"ashborn.id/moniplan/middlewares"
	"github.com/gin-gonic/gin"
)

// SetupRoutes mengkonfigurasi semua routes aplikasi
func SetupRoutes(router *gin.Engine) {
	// Health check endpoint - selalu public
	router.GET("/health", controllers.HealthCheck)

	// API version grouping
	v1 := router.Group("/api/v1")
	{
		// Public auth routes (tidak perlu authentication)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// Protected routes (perlu authentication)
		protected := v1.Group("/")
		protected.Use(middlewares.AuthMiddleware()) // Apply auth middleware
		{
			// User routes
			protected.GET("/profile", controllers.GetProfile)
			protected.POST("/auth/refresh", controllers.RefreshToken)

			// Category routes
			protected.GET("/category", controllers.IndexCategory)
			protected.POST("/category/create", controllers.CreateCategory)
			protected.GET("/category/:id", controllers.GetCategoryByID)
			protected.POST("/category/update/:id", controllers.UpdateCategory)
			protected.GET("/category/delete/:id", controllers.DeleteCategoryByID)

			// Transaction routes
			protected.GET("/transaction", controllers.IndexTransaction)
			protected.POST("/transaction/create", controllers.CreateTransaction)
			protected.GET("/transaction/:id", controllers.GetTransactionByID)
			protected.POST("/transaction/update/:id", controllers.UpdateTransaction)
			protected.GET("/transaction/delete/:id", controllers.DeleteTransactionByID)
		}
	}

	// Setup 404 handler untuk undefined routes
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":   "Not Found",
			"message": "The requested endpoint does not exist",
		})
	})
}

// SetupMiddlewares mengkonfigurasi global middlewares
func SetupMiddlewares(router *gin.Engine) {
	// Recovery middleware untuk handle panic
	router.Use(gin.Recovery())

	// Logger middleware untuk log setiap request
	router.Use(gin.Logger())

	// CORS middleware untuk allow cross-origin requests
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

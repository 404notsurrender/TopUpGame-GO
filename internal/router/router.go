package router

import (
	"topup-game/internal/handler"
	"topup-game/internal/middleware"
	"topup-game/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(
	userService service.UserService,
	productService service.ProductService,
	transactionService service.TransactionService,
) *gin.Engine {
	router := gin.New()

	// Use logger and recovery middleware
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorLogger())
	router.Use(middleware.RecoveryLogger())

	// Create handlers
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	// Create auth middlewares
	authMiddleware := middleware.AuthMiddleware(userService)
	adminMiddleware := middleware.AdminMiddleware(userService)
	optionalAuthMiddleware := middleware.OptionalAuthMiddleware(userService)

	// Static files
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// Public routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "Top Up Game - Home",
		})
	})

	// Register routes for each handler
	userHandler.RegisterRoutes(router)
	productHandler.RegisterRoutes(router, adminMiddleware)
	transactionHandler.RegisterRoutes(router, authMiddleware, adminMiddleware)

	// API routes
	api := router.Group("/api")
	{
		// Public endpoints
		api.GET("/products", productHandler.ListProducts)
		api.GET("/products/:id", productHandler.GetProduct)
		api.POST("/checkout", optionalAuthMiddleware, transactionHandler.Checkout)
		api.GET("/transaction/:invoice", transactionHandler.GetTransactionStatus)

		// Auth endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/register", userHandler.Register)
		}

		// Protected endpoints
		protected := api.Group("/user")
		protected.Use(authMiddleware)
		{
			protected.GET("/profile", userHandler.GetProfile)
			protected.GET("/transactions", transactionHandler.GetUserTransactions)
		}

		// Admin endpoints
		admin := api.Group("/admin")
		admin.Use(authMiddleware, adminMiddleware)
		{
			// Product management
			admin.POST("/products", productHandler.CreateProduct)
			admin.PUT("/products/:id", productHandler.UpdateProduct)
			admin.DELETE("/products/:id", productHandler.DeleteProduct)
			admin.POST("/products/sync", productHandler.SyncProducts)

			// Transaction management
			admin.GET("/transactions", transactionHandler.ListTransactions)
			admin.GET("/transactions/:id", transactionHandler.GetTransaction)
		}
	}

	return router
}

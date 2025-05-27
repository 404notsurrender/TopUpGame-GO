package main

import (
	"fmt"
	"log"
	"topup-game/config"
	"topup-game/internal/model"
	"topup-game/internal/repository"
	"topup-game/internal/router"
	"topup-game/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Auto migrate database
	err = cfg.DB.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.Transaction{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(cfg.DB)
	productRepo := repository.NewProductRepository(cfg.DB)
	transactionRepo := repository.NewTransactionRepository(cfg.DB)

	// Initialize VIP Reseller service
	vipResellerService := service.NewVIPResellerService(
		cfg.VIPReseller.BaseURL,
		cfg.VIPReseller.APIKey,
		cfg.VIPReseller.UserID,
	)

	// Initialize services
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	productService := service.NewProductService(productRepo, vipResellerService)
	transactionService := service.NewTransactionService(transactionRepo, productRepo, vipResellerService)

	// Setup router
	r := router.SetupRouter(userService, productService, transactionService)

	// Create default admin user if not exists
	createDefaultAdmin(userService)

	// Start server
	port := "8080"
	fmt.Printf("Server is running on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createDefaultAdmin(userService service.UserService) {
	_, err := userService.Register("admin@example.com", "admin123", model.RoleAdmin)
	if err != nil {
		// If error is not because email is taken, log it
		if err != repository.ErrEmailTaken {
			log.Printf("Failed to create default admin: %v", err)
		}
	}
}

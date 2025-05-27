package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds all configuration for our application
type Config struct {
	DB           *gorm.DB
	JWTSecret    string
	VIPReseller  VIPResellerConfig
}

// VIPResellerConfig holds configuration for VIP Reseller API
type VIPResellerConfig struct {
	APIKey  string
	UserID  string
	BaseURL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	// Database connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	// Initialize database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Return config instance
	return &Config{
		DB:        db,
		JWTSecret: os.Getenv("JWT_SECRET"),
		VIPReseller: VIPResellerConfig{
			APIKey:  os.Getenv("VIP_RESELLER_API_KEY"),
			UserID:  os.Getenv("VIP_RESELLER_USER_ID"),
			BaseURL: os.Getenv("VIP_RESELLER_BASE_URL"),
		},
	}, nil
}

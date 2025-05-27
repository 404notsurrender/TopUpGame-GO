package model

import (
	"time"

	"gorm.io/gorm"
)

// Product represents the product model for game top-up items
type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Category    string         `gorm:"not null;index" json:"category"`
	Price       float64        `gorm:"not null" json:"price"`
	Description string         `gorm:"type:text" json:"description"`
	SKU         string         `gorm:"uniqueIndex" json:"sku"`
	IsActive    bool          `gorm:"default:true" json:"is_active"`
	Stock       int           `gorm:"default:0" json:"stock"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Product model
func (Product) TableName() string {
	return "products"
}

// BeforeCreate hook is called before creating a new product record
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-creation logic here if needed
	return nil
}

// Validate performs validation on product data
func (p *Product) Validate() error {
	if p.Name == "" {
		return ErrProductNameRequired
	}
	if p.Category == "" {
		return ErrProductCategoryRequired
	}
	if p.Price <= 0 {
		return ErrInvalidProductPrice
	}
	return nil
}

// Custom errors for product validation
var (
	ErrProductNameRequired     = ValidationError{"product name is required"}
	ErrProductCategoryRequired = ValidationError{"product category is required"}
	ErrInvalidProductPrice     = ValidationError{"product price must be greater than 0"}
)

// ValidationError represents a validation error
type ValidationError struct {
	message string
}

func (e ValidationError) Error() string {
	return e.message
}

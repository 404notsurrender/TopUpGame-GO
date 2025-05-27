package model

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleGuest    Role = "guest"
	RoleAdmin    Role = "admin"
	RoleReseller Role = "reseller"
)

// User represents the user model
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Role      Role          `gorm:"type:varchar(10);not null" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook is called before creating a new user record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-creation logic here if needed
	return nil
}

// Validate performs validation on user data
func (u *User) Validate() error {
	// Add validation logic here if needed
	return nil
}

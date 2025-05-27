package model

import (
	"time"

	"gorm.io/gorm"
)

type TransactionStatus string

const (
	StatusPending TransactionStatus = "pending"
	StatusSuccess TransactionStatus = "success"
	StatusFailed  TransactionStatus = "failed"
)

// Transaction represents the transaction model for game top-up purchases
type Transaction struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	UserID        *uint             `json:"user_id"`
	User          *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ProductID     uint              `json:"product_id"`
	Product       Product           `gorm:"foreignKey:ProductID" json:"product"`
	Method        string            `gorm:"not null" json:"method"`
	Invoice       string            `gorm:"uniqueIndex;not null" json:"invoice"`
	Status        TransactionStatus `gorm:"type:varchar(10);not null;index" json:"status"`
	Amount        float64           `gorm:"not null" json:"amount"`
	GameID        string            `gorm:"not null" json:"game_id"`
	GameServer    string            `gorm:"not null" json:"game_server"`
	PaymentProof  string            `gorm:"type:text" json:"payment_proof,omitempty"`
	Notes         string            `gorm:"type:text" json:"notes,omitempty"`
	VipOrderID    string            `gorm:"uniqueIndex" json:"vip_order_id,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Transaction model
func (Transaction) TableName() string {
	return "transactions"
}

// BeforeCreate hook is called before creating a new transaction record
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.Status == "" {
		t.Status = StatusPending
	}
	return nil
}

// Validate performs validation on transaction data
func (t *Transaction) Validate() error {
	if t.ProductID == 0 {
		return ErrProductIDRequired
	}
	if t.Method == "" {
		return ErrPaymentMethodRequired
	}
	if t.GameID == "" {
		return ErrGameIDRequired
	}
	if t.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// IsComplete checks if the transaction is in a final state
func (t *Transaction) IsComplete() bool {
	return t.Status == StatusSuccess || t.Status == StatusFailed
}

// Custom errors for transaction validation
var (
	ErrProductIDRequired     = ValidationError{"product ID is required"}
	ErrPaymentMethodRequired = ValidationError{"payment method is required"}
	ErrGameIDRequired        = ValidationError{"game ID is required"}
	ErrInvalidAmount        = ValidationError{"amount must be greater than 0"}
)

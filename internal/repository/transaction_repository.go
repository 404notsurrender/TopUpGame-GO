package repository

import (
	"errors"
	"topup-game/internal/model"

	"gorm.io/gorm"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvoiceExists      = errors.New("invoice number already exists")
)

type TransactionRepository interface {
	Create(transaction *model.Transaction) error
	Update(transaction *model.Transaction) error
	FindByID(id uint) (*model.Transaction, error)
	FindByInvoice(invoice string) (*model.Transaction, error)
	FindByVipOrderID(orderID string) (*model.Transaction, error)
	FindAll(params TransactionQueryParams) ([]model.Transaction, error)
	FindByUserID(userID uint) ([]model.Transaction, error)
	UpdateStatus(id uint, status model.TransactionStatus) error
}

type TransactionQueryParams struct {
	Status    *model.TransactionStatus
	Method    string
	StartDate string
	EndDate   string
	Search    string
	SortBy    string
	SortDesc  bool
	Limit     int
	Offset    int
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(transaction *model.Transaction) error {
	// Check if invoice already exists
	var count int64
	if err := r.db.Model(&model.Transaction{}).Where("invoice = ?", transaction.Invoice).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrInvoiceExists
	}

	return r.db.Create(transaction).Error
}

func (r *transactionRepository) Update(transaction *model.Transaction) error {
	result := r.db.Save(transaction)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTransactionNotFound
	}
	return nil
}

func (r *transactionRepository) FindByID(id uint) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.db.Preload("Product").Preload("User").First(&transaction, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) FindByInvoice(invoice string) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.db.Preload("Product").Preload("User").
		Where("invoice = ?", invoice).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) FindByVipOrderID(orderID string) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.db.Preload("Product").Preload("User").
		Where("vip_order_id = ?", orderID).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) FindAll(params TransactionQueryParams) ([]model.Transaction, error) {
	var transactions []model.Transaction
	query := r.db.Preload("Product").Preload("User")

	// Apply filters
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.Method != "" {
		query = query.Where("method = ?", params.Method)
	}
	if params.StartDate != "" {
		query = query.Where("created_at >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("created_at <= ?", params.EndDate)
	}
	if params.Search != "" {
		query = query.Where("invoice ILIKE ? OR game_id ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	// Apply sorting
	if params.SortBy != "" {
		direction := "ASC"
		if params.SortDesc {
			direction = "DESC"
		}
		query = query.Order(params.SortBy + " " + direction)
	} else {
		// Default sort by created_at desc
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	err := query.Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) FindByUserID(userID uint) ([]model.Transaction, error) {
	var transactions []model.Transaction
	err := r.db.Preload("Product").Where("user_id = ?", userID).
		Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) UpdateStatus(id uint, status model.TransactionStatus) error {
	result := r.db.Model(&model.Transaction{}).Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTransactionNotFound
	}
	return nil
}

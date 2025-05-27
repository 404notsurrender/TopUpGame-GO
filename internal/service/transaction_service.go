package service

import (
	"errors"
	"fmt"
	"time"
	"topup-game/internal/model"
	"topup-game/internal/repository"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction data")
	ErrProductUnavailable = errors.New("product is currently unavailable")
)

type TransactionService interface {
	CreateTransaction(transaction *model.Transaction) error
	GetTransactionByID(id uint) (*model.Transaction, error)
	GetTransactionByInvoice(invoice string) (*model.Transaction, error)
	GetTransactions(params repository.TransactionQueryParams) ([]model.Transaction, error)
	GetUserTransactions(userID uint) ([]model.Transaction, error)
	UpdateTransactionStatus(id uint, status model.TransactionStatus) error
	ProcessCheckout(checkout CheckoutRequest) (*model.Transaction, error)
	SyncTransactionStatus(invoice string) error
}

type CheckoutRequest struct {
	ProductID   uint   `json:"product_id" validate:"required"`
	UserID      *uint  `json:"user_id"`
	GameID      string `json:"game_id" validate:"required"`
	GameServer  string `json:"game_server" validate:"required"`
	Method      string `json:"method" validate:"required"`
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	productRepo     repository.ProductRepository
	vipReseller     VIPResellerService
}

func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	productRepo repository.ProductRepository,
	vipReseller VIPResellerService,
) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		vipReseller:     vipReseller,
	}
}

func (s *transactionService) CreateTransaction(transaction *model.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTransaction, err)
	}

	// Generate unique invoice number
	transaction.Invoice = generateInvoiceNumber()

	// Set initial status
	transaction.Status = model.StatusPending

	// Create transaction
	return s.transactionRepo.Create(transaction)
}

func (s *transactionService) GetTransactionByID(id uint) (*model.Transaction, error) {
	return s.transactionRepo.FindByID(id)
}

func (s *transactionService) GetTransactionByInvoice(invoice string) (*model.Transaction, error) {
	return s.transactionRepo.FindByInvoice(invoice)
}

func (s *transactionService) GetTransactions(params repository.TransactionQueryParams) ([]model.Transaction, error) {
	return s.transactionRepo.FindAll(params)
}

func (s *transactionService) GetUserTransactions(userID uint) ([]model.Transaction, error) {
	return s.transactionRepo.FindByUserID(userID)
}

func (s *transactionService) UpdateTransactionStatus(id uint, status model.TransactionStatus) error {
	return s.transactionRepo.UpdateStatus(id, status)
}

func (s *transactionService) ProcessCheckout(checkout CheckoutRequest) (*model.Transaction, error) {
	// Get product details
	product, err := s.productRepo.FindByID(checkout.ProductID)
	if err != nil {
		return nil, err
	}

	// Check if product is available
	if !product.IsActive || product.Stock <= 0 {
		return nil, ErrProductUnavailable
	}

	// Create transaction
	transaction := &model.Transaction{
		UserID:     checkout.UserID,
		ProductID:  checkout.ProductID,
		Method:     checkout.Method,
		Amount:     product.Price,
		GameID:     checkout.GameID,
		GameServer: checkout.GameServer,
		Status:     model.StatusPending,
	}

	// Save transaction
	if err := s.CreateTransaction(transaction); err != nil {
		return nil, err
	}

	// Create order in VIP Reseller
	vipOrder := VIPOrder{
		GameID:     checkout.GameID,
		GameServer: checkout.GameServer,
		ProductSKU: product.SKU,
	}

	vipResponse, err := s.vipReseller.CreateOrder(vipOrder)
	if err != nil {
		// Update transaction status to failed if VIP Reseller order fails
		_ = s.UpdateTransactionStatus(transaction.ID, model.StatusFailed)
		return nil, fmt.Errorf("failed to create VIP Reseller order: %v", err)
	}

	// Update transaction with VIP order ID
	transaction.VipOrderID = vipResponse.OrderID
	if err := s.transactionRepo.Update(transaction); err != nil {
		return nil, err
	}

	// Decrease product stock
	if err := s.productRepo.UpdateStock(product.ID, -1); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *transactionService) SyncTransactionStatus(invoice string) error {
	// Get transaction
	transaction, err := s.transactionRepo.FindByInvoice(invoice)
	if err != nil {
		return err
	}

	// Skip if transaction is already in final state
	if transaction.IsComplete() {
		return nil
	}

	// Check status in VIP Reseller
	if transaction.VipOrderID != "" {
		vipStatus, err := s.vipReseller.CheckStatus(transaction.VipOrderID)
		if err != nil {
			return fmt.Errorf("failed to check VIP Reseller status: %v", err)
		}

		// Map VIP Reseller status to our status
		var newStatus model.TransactionStatus
		switch vipStatus.Status {
		case "success":
			newStatus = model.StatusSuccess
		case "failed":
			newStatus = model.StatusFailed
		default:
			newStatus = model.StatusPending
		}

		// Update transaction status if changed
		if transaction.Status != newStatus {
			if err := s.UpdateTransactionStatus(transaction.ID, newStatus); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper function to generate unique invoice number
func generateInvoiceNumber() string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("INV/%s/%d", timestamp, time.Now().UnixNano()%1000)
}

package handler

import (
	"net/http"
	"strconv"
	"topup-game/internal/model"
	"topup-game/internal/repository"
	"topup-game/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TransactionHandler struct {
	transactionService service.TransactionService
	validator          *validator.Validate
}

func NewTransactionHandler(transactionService service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		validator:          validator.New(),
	}
}

type CheckoutRequest struct {
	ProductID   uint   `json:"product_id" validate:"required"`
	GameID      string `json:"game_id" validate:"required"`
	GameServer  string `json:"game_server" validate:"required"`
	Method      string `json:"method" validate:"required,oneof=bank_transfer ewallet credit_card"`
}

// Checkout handles creating a new transaction
func (h *TransactionHandler) Checkout(c *gin.Context) {
	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Get user ID from context if authenticated
	var userID *uint
	if id, exists := c.Get("userID"); exists {
		uid := id.(uint)
		userID = &uid
	}

	checkout := service.CheckoutRequest{
		ProductID:  req.ProductID,
		UserID:     userID,
		GameID:     req.GameID,
		GameServer: req.GameServer,
		Method:     req.Method,
	}

	transaction, err := h.transactionService.ProcessCheckout(checkout)
	if err != nil {
		switch err {
		case service.ErrProductUnavailable:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product is currently unavailable"})
		case service.ErrInvalidTransaction:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction data"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process checkout"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Checkout successful",
		"transaction": gin.H{
			"id":      transaction.ID,
			"invoice": transaction.Invoice,
			"amount":  transaction.Amount,
			"status":  transaction.Status,
		},
	})
}

// GetTransactionStatus handles fetching transaction status by invoice
func (h *TransactionHandler) GetTransactionStatus(c *gin.Context) {
	invoice := c.Param("invoice")
	if invoice == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice number is required"})
		return
	}

	// Sync status with VIP Reseller
	if err := h.transactionService.SyncTransactionStatus(invoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync transaction status"})
		return
	}

	// Get updated transaction
	transaction, err := h.transactionService.GetTransactionByInvoice(invoice)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction": gin.H{
			"id":         transaction.ID,
			"invoice":    transaction.Invoice,
			"amount":     transaction.Amount,
			"status":     transaction.Status,
			"game_id":    transaction.GameID,
			"created_at": transaction.CreatedAt,
			"updated_at": transaction.UpdatedAt,
		},
	})
}

// ListTransactions handles fetching all transactions (admin only)
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	params := repository.TransactionQueryParams{
		Search:    c.Query("search"),
		Method:    c.Query("method"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortDesc:  c.DefaultQuery("sort_dir", "desc") == "desc",
	}

	if status := c.Query("status"); status != "" {
		transStatus := model.TransactionStatus(status)
		params.Status = &transStatus
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			params.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			params.Offset = o
		}
	}

	transactions, err := h.transactionService.GetTransactions(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// GetTransaction handles fetching a single transaction (admin only)
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	transaction, err := h.transactionService.GetTransactionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": transaction})
}

// GetUserTransactions handles fetching transactions for the authenticated user
func (h *TransactionHandler) GetUserTransactions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	transactions, err := h.transactionService.GetUserTransactions(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// RegisterRoutes registers the transaction routes
func (h *TransactionHandler) RegisterRoutes(router *gin.Engine, authMiddleware, adminMiddleware gin.HandlerFunc) {
	// Public routes
	router.POST("/checkout", h.Checkout)
	router.GET("/transaction/:invoice", h.GetTransactionStatus)

	// Protected routes
	protected := router.Group("/transactions")
	protected.Use(authMiddleware)
	{
		protected.GET("/my", h.GetUserTransactions)
	}

	// Admin routes
	admin := router.Group("/admin/transactions")
	admin.Use(adminMiddleware)
	{
		admin.GET("", h.ListTransactions)
		admin.GET("/:id", h.GetTransaction)
	}
}

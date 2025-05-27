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

type ProductHandler struct {
	productService service.ProductService
	validator      *validator.Validate
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		validator:      validator.New(),
	}
}

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Description string  `json:"description"`
	SKU         string  `json:"sku" validate:"required"`
	Stock       int     `json:"stock" validate:"min=0"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Description string  `json:"description"`
	IsActive    bool    `json:"is_active"`
}

// ListProducts handles fetching all products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// Parse query parameters
	params := repository.ProductQueryParams{
		Category: c.Query("category"),
		Search:   c.Query("search"),
		SortBy:   c.DefaultQuery("sort_by", "created_at"),
		SortDesc: c.DefaultQuery("sort_dir", "desc") == "desc",
	}

	if active := c.Query("active"); active != "" {
		isActive := active == "true"
		params.IsActive = &isActive
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

	products, err := h.productService.GetProducts(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// CreateProduct handles creating a new product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	product := &model.Product{
		Name:        req.Name,
		Category:    req.Category,
		Price:       req.Price,
		Description: req.Description,
		SKU:         req.SKU,
		Stock:       req.Stock,
		IsActive:    true,
	}

	if err := h.productService.CreateProduct(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"product": product})
}

// UpdateProduct handles updating an existing product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	product := &model.Product{
		ID:          uint(id),
		Name:        req.Name,
		Category:    req.Category,
		Price:       req.Price,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := h.productService.UpdateProduct(product); err != nil {
		if err == service.ErrInvalidProduct {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

// GetProduct handles fetching a single product
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := h.productService.GetProductByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

// DeleteProduct handles deleting a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := h.productService.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// SyncProducts handles syncing products with VIP Reseller
func (h *ProductHandler) SyncProducts(c *gin.Context) {
	if err := h.productService.SyncProductsWithVIPReseller(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Products synced successfully"})
}

// RegisterRoutes registers the product routes
func (h *ProductHandler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// Public routes
	router.GET("/products", h.ListProducts)
	router.GET("/products/:id", h.GetProduct)

	// Admin routes
	admin := router.Group("/admin/products")
	admin.Use(authMiddleware)
	{
		admin.POST("", h.CreateProduct)
		admin.PUT("/:id", h.UpdateProduct)
		admin.DELETE("/:id", h.DeleteProduct)
		admin.POST("/sync", h.SyncProducts)
	}
}

package service

import (
	"errors"
	"fmt"
	"topup-game/internal/model"
	"topup-game/internal/repository"
)

var (
	ErrInvalidProduct = errors.New("invalid product data")
	ErrOutOfStock    = errors.New("product is out of stock")
)

type ProductService interface {
	CreateProduct(product *model.Product) error
	UpdateProduct(product *model.Product) error
	DeleteProduct(id uint) error
	GetProductByID(id uint) (*model.Product, error)
	GetProducts(params repository.ProductQueryParams) ([]model.Product, error)
	GetProductsByCategory(category string) ([]model.Product, error)
	UpdateStock(id uint, quantity int) error
	SyncProductsWithVIPReseller() error
}

type productService struct {
	productRepo repository.ProductRepository
	vipReseller VIPResellerService
}

func NewProductService(productRepo repository.ProductRepository, vipReseller VIPResellerService) ProductService {
	return &productService{
		productRepo: productRepo,
		vipReseller: vipReseller,
	}
}

func (s *productService) CreateProduct(product *model.Product) error {
	if err := product.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidProduct, err)
	}

	return s.productRepo.Create(product)
}

func (s *productService) UpdateProduct(product *model.Product) error {
	if err := product.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidProduct, err)
	}

	// Check if product exists
	existing, err := s.productRepo.FindByID(product.ID)
	if err != nil {
		return err
	}

	// Update only allowed fields
	existing.Name = product.Name
	existing.Category = product.Category
	existing.Price = product.Price
	existing.Description = product.Description
	existing.IsActive = product.IsActive
	existing.SKU = product.SKU

	return s.productRepo.Update(existing)
}

func (s *productService) DeleteProduct(id uint) error {
	return s.productRepo.Delete(id)
}

func (s *productService) GetProductByID(id uint) (*model.Product, error) {
	return s.productRepo.FindByID(id)
}

func (s *productService) GetProducts(params repository.ProductQueryParams) ([]model.Product, error) {
	return s.productRepo.FindAll(params)
}

func (s *productService) GetProductsByCategory(category string) ([]model.Product, error) {
	return s.productRepo.FindByCategory(category)
}

func (s *productService) UpdateStock(id uint, quantity int) error {
	// Get current product
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Check if we have enough stock for negative quantities
	if quantity < 0 && (product.Stock + quantity) < 0 {
		return ErrOutOfStock
	}

	return s.productRepo.UpdateStock(id, quantity)
}

func (s *productService) SyncProductsWithVIPReseller() error {
	// Get products from VIP Reseller API
	vipProducts, err := s.vipReseller.GetGameFeatures()
	if err != nil {
		return fmt.Errorf("failed to fetch VIP Reseller products: %v", err)
	}

	// For each VIP Reseller product, create or update in our database
	for _, vipProduct := range vipProducts {
		product, err := s.productRepo.FindBySKU(vipProduct.SKU)
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				// Create new product
				newProduct := &model.Product{
					Name:        vipProduct.Name,
					Category:    vipProduct.Category,
					Price:      vipProduct.Price,
					Description: vipProduct.Description,
					SKU:        vipProduct.SKU,
					IsActive:   true,
					Stock:      vipProduct.Stock,
				}
				if err := s.CreateProduct(newProduct); err != nil {
					return fmt.Errorf("failed to create product %s: %v", vipProduct.SKU, err)
				}
			} else {
				return fmt.Errorf("error checking product %s: %v", vipProduct.SKU, err)
			}
		} else {
			// Update existing product
			product.Name = vipProduct.Name
			product.Price = vipProduct.Price
			product.Description = vipProduct.Description
			product.Stock = vipProduct.Stock
			if err := s.UpdateProduct(product); err != nil {
				return fmt.Errorf("failed to update product %s: %v", vipProduct.SKU, err)
			}
		}
	}

	return nil
}

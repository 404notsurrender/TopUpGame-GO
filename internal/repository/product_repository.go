package repository

import (
	"errors"
	"topup-game/internal/model"

	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrSKUTaken       = errors.New("product SKU already taken")
)

type ProductRepository interface {
	Create(product *model.Product) error
	Update(product *model.Product) error
	Delete(id uint) error
	FindByID(id uint) (*model.Product, error)
	FindBySKU(sku string) (*model.Product, error)
	FindAll(params ProductQueryParams) ([]model.Product, error)
	FindByCategory(category string) ([]model.Product, error)
	UpdateStock(id uint, quantity int) error
}

type ProductQueryParams struct {
	Category string
	IsActive *bool
	Search   string
	SortBy   string
	SortDesc bool
	Limit    int
	Offset   int
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *model.Product) error {
	// Check if SKU already exists
	if product.SKU != "" {
		var count int64
		if err := r.db.Model(&model.Product{}).Where("sku = ?", product.SKU).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrSKUTaken
		}
	}

	return r.db.Create(product).Error
}

func (r *productRepository) Update(product *model.Product) error {
	result := r.db.Save(product)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *productRepository) Delete(id uint) error {
	result := r.db.Delete(&model.Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *productRepository) FindByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindBySKU(sku string) (*model.Product, error) {
	var product model.Product
	err := r.db.Where("sku = ?", sku).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindAll(params ProductQueryParams) ([]model.Product, error) {
	var products []model.Product
	query := r.db

	// Apply filters
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}
	if params.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	// Apply sorting
	if params.SortBy != "" {
		direction := "ASC"
		if params.SortDesc {
			direction = "DESC"
		}
		query = query.Order(params.SortBy + " " + direction)
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	err := query.Find(&products).Error
	return products, err
}

func (r *productRepository) FindByCategory(category string) ([]model.Product, error) {
	var products []model.Product
	err := r.db.Where("category = ? AND is_active = true", category).Find(&products).Error
	return products, err
}

func (r *productRepository) UpdateStock(id uint, quantity int) error {
	result := r.db.Model(&model.Product{}).Where("id = ?", id).
		Update("stock", gorm.Expr("stock + ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

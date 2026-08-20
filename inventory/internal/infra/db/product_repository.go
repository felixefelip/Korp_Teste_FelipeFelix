package db

import (
	"inventory/internal/model"

	"gorm.io/gorm"
)

type ProductRepository struct {
	connection *gorm.DB
}

func NewProductRepository(connection *gorm.DB) *ProductRepository {
	return &ProductRepository{
		connection: connection,
	}
}

func (pr *ProductRepository) GetProducts() ([]model.Product, error) {
	var productList []model.Product

	err := pr.connection.Find(&productList).Error
	if err != nil {
		return []model.Product{}, err
	}

	return productList, nil
}

func (pr *ProductRepository) GetProductByID(id int) (model.Product, error) {
	var product model.Product

	err := pr.connection.First(&product, id).Error
	if err != nil {
		return model.Product{}, err
	}

	return product, nil
}

func (pr *ProductRepository) CreateProduct(product model.Product) (int, error) {
	err := pr.connection.Create(&product).Error
	if err != nil {
		return 0, err
	}

	return product.ID, nil
}

func (pr *ProductRepository) UpdateProduct(product model.Product) error {
	result := pr.connection.
		Model(&model.Product{ID: product.ID}).
		Select("code", "name", "unit", "price", "stock").
		Updates(product)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

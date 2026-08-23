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
	err := pr.connection.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			return err
		}

		event, err := model.NewProductCreated(product)
		if err != nil {
			return err
		}

		return tx.Create(&event).Error
	})
	if err != nil {
		return 0, err
	}

	return product.ID, nil
}

func (pr *ProductRepository) UpdateProduct(product model.Product) error {
	return pr.connection.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&model.Product{ID: product.ID}).
			Select("code", "name", "unit", "price").
			Updates(product)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		event, err := model.NewProductUpdated(product)
		if err != nil {
			return err
		}

		return tx.Create(&event).Error
	})
}

func (pr *ProductRepository) DeleteProduct(id int) error {
	return pr.connection.Transaction(func(tx *gorm.DB) error {
		if err := deleteStockMovementsByProductID(tx, id); err != nil {
			return err
		}

		result := tx.Delete(&model.Product{}, id)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		event, err := model.NewProductDeleted(id)
		if err != nil {
			return err
		}

		return tx.Create(&event).Error
	})
}

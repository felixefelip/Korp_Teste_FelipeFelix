package db

import (
	"billing/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	var products []model.Product

	err := pr.connection.
		Where("active").
		Order("code").
		Find(&products).Error
	if err != nil {
		return []model.Product{}, err
	}

	return products, nil
}

func (pr *ProductRepository) UpsertProduct(product model.Product) error {
	product.Active = true

	return pr.connection.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "inventory_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"code", "name", "unit", "price", "active"}),
	}).Create(&product).Error
}

func (pr *ProductRepository) DeactivateProduct(inventoryID int) error {
	return pr.connection.
		Model(&model.Product{}).
		Where("inventory_id = ?", inventoryID).
		Update("active", false).Error
}

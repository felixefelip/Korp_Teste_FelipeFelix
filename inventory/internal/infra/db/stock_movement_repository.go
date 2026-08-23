package db

import (
	"inventory/internal/model"

	"gorm.io/gorm"
)

type StockMovementRepository struct {
	connection *gorm.DB
}

func NewStockMovementRepository(connection *gorm.DB) *StockMovementRepository {
	return &StockMovementRepository{
		connection: connection,
	}
}

func (sr *StockMovementRepository) GetMovementsByProductID(productID int) ([]model.StockMovement, error) {
	var movements []model.StockMovement

	err := sr.connection.
		Where("product_id = ?", productID).
		Order("id desc").
		Find(&movements).Error
	if err != nil {
		return []model.StockMovement{}, err
	}

	return movements, nil
}

func (sr *StockMovementRepository) GetMovementByID(id int) (model.StockMovement, error) {
	var movement model.StockMovement

	err := sr.connection.First(&movement, id).Error
	if err != nil {
		return model.StockMovement{}, err
	}

	return movement, nil
}

func (sr *StockMovementRepository) CreateMovement(movement model.StockMovement) (int, error) {
	err := sr.connection.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&movement).Error; err != nil {
			return err
		}

		return refreshStock(tx, movement.ProductID)
	})
	if err != nil {
		return 0, err
	}

	return movement.ID, nil
}

func (sr *StockMovementRepository) UpdateMovement(movement model.StockMovement) error {
	return sr.connection.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&model.StockMovement{ID: movement.ID}).
			Select("type", "quantity", "confirmed").
			Updates(movement)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return refreshStock(tx, movement.ProductID)
	})
}

func (sr *StockMovementRepository) ApplyInvoice(request model.InvoiceStockRequest, event model.OutboxEvent) error {
	return sr.connection.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&event).Error
	})
}

func deleteStockMovementsByProductID(tx *gorm.DB, productID int) error {
	return tx.Where("product_id = ?", productID).Delete(&model.StockMovement{}).Error
}

func refreshStock(tx *gorm.DB, productID int) error {
	return tx.Exec(`
		UPDATE product
		SET stock = COALESCE((
			SELECT SUM(CASE WHEN type = ? THEN quantity ELSE -quantity END)
			FROM stock_movement
			WHERE product_id = product.id AND confirmed
		), 0)
		WHERE id = ?`, model.MovementIn, productID).Error
}

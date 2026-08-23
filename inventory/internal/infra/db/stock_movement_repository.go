package db

import (
	"inventory/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (sr *StockMovementRepository) ApplyInvoice(
	request model.InvoiceStockRequest,
) (model.OutboxEvent, error) {
	var decision model.InvoiceStockDecision

	err := sr.connection.Transaction(func(tx *gorm.DB) error {
		products, err := lockProducts(tx, request.ProductIDs())
		if err != nil {
			return err
		}

		applied, err := alreadyApplied(tx, request.InvoiceID)
		if err != nil {
			return err
		}

		decision, err = model.ResolveInvoiceStock(request, products, applied)
		if err != nil {
			return err
		}

		if err := writeMovements(tx, decision.Movements); err != nil {
			return err
		}

		return tx.Create(&decision.Event).Error
	})
	if err != nil {
		return model.OutboxEvent{}, err
	}

	return decision.Event, nil
}

func (sr *StockMovementRepository) RevertInvoice(
	request model.InvoiceStockRevertRequest,
) (model.OutboxEvent, error) {
	var decision model.InvoiceRevertDecision

	err := sr.connection.Transaction(func(tx *gorm.DB) error {
		productIDs, err := productsTouchedByInvoice(tx, request.InvoiceID)
		if err != nil {
			return err
		}

		products, err := lockProducts(tx, productIDs)
		if err != nil {
			return err
		}

		movements, err := movementsOfInvoice(tx, request.InvoiceID)
		if err != nil {
			return err
		}

		decision, err = model.ResolveInvoiceRevert(request, movements, products)
		if err != nil {
			return err
		}

		if err := eraseMovements(tx, decision.Movements); err != nil {
			return err
		}

		return tx.Create(&decision.Event).Error
	})
	if err != nil {
		return model.OutboxEvent{}, err
	}

	return decision.Event, nil
}

func productsTouchedByInvoice(tx *gorm.DB, invoiceID int) ([]int, error) {
	var ids []int

	err := tx.
		Model(&model.StockMovement{}).
		Where("billing_invoice_id = ?", invoiceID).
		Distinct().
		Order("product_id").
		Pluck("product_id", &ids).Error

	return ids, err
}

func movementsOfInvoice(tx *gorm.DB, invoiceID int) ([]model.StockMovement, error) {
	var movements []model.StockMovement

	err := tx.
		Where("billing_invoice_id = ?", invoiceID).
		Order("id").
		Find(&movements).Error

	return movements, err
}

func eraseMovements(tx *gorm.DB, movements []model.StockMovement) error {
	if len(movements) == 0 {
		return nil
	}

	ids := make([]int, 0, len(movements))
	touched := make(map[int]struct{}, len(movements))

	for _, movement := range movements {
		ids = append(ids, movement.ID)
		touched[movement.ProductID] = struct{}{}
	}

	if err := tx.Where("id IN ?", ids).Delete(&model.StockMovement{}).Error; err != nil {
		return err
	}

	for productID := range touched {
		if err := refreshStock(tx, productID); err != nil {
			return err
		}
	}

	return nil
}

func lockProducts(tx *gorm.DB, ids []int) (map[int]model.Product, error) {
	locked := make(map[int]model.Product, len(ids))

	if len(ids) == 0 {
		return locked, nil
	}

	var products []model.Product

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", ids).
		Order("id").
		Find(&products).Error
	if err != nil {
		return nil, err
	}

	for _, product := range products {
		locked[product.ID] = product
	}

	return locked, nil
}

func alreadyApplied(tx *gorm.DB, invoiceID int) (bool, error) {
	var count int64

	err := tx.
		Model(&model.StockMovement{}).
		Where("billing_invoice_id = ?", invoiceID).
		Count(&count).Error

	return count > 0, err
}

func writeMovements(tx *gorm.DB, movements []model.StockMovement) error {
	if len(movements) == 0 {
		return nil
	}

	touched := make(map[int]struct{}, len(movements))

	for _, movement := range movements {
		if err := tx.Create(&movement).Error; err != nil {
			return err
		}

		touched[movement.ProductID] = struct{}{}
	}

	for productID := range touched {
		if err := refreshStock(tx, productID); err != nil {
			return err
		}
	}

	return nil
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

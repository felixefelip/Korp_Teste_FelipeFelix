package db

import (
	"billing/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InvoiceRepository struct {
	connection *gorm.DB
}

func NewInvoiceRepository(connection *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{
		connection: connection,
	}
}

func (ir *InvoiceRepository) GetInvoices() ([]model.Invoice, error) {
	var invoiceList []model.Invoice

	err := withItems(ir.connection).Find(&invoiceList).Error
	if err != nil {
		return []model.Invoice{}, err
	}

	return invoiceList, nil
}

func (ir *InvoiceRepository) GetInvoiceByID(id int) (model.Invoice, error) {
	var invoice model.Invoice

	err := withItems(ir.connection).First(&invoice, id).Error
	if err != nil {
		return model.Invoice{}, err
	}

	return invoice, nil
}

func (ir *InvoiceRepository) CreateInvoice(invoice model.Invoice) (int, error) {
	err := ir.connection.Transaction(func(tx *gorm.DB) error {
		items := invoice.Items
		invoice.Items = nil

		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}

		return saveItems(tx, invoice.ID, items)
	})
	if err != nil {
		return 0, err
	}

	return invoice.ID, nil
}

func (ir *InvoiceRepository) UpdateInvoice(invoice model.Invoice) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&model.Invoice{ID: invoice.ID}).
			Select("number").
			Updates(model.Invoice{Number: invoice.Number})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if invoice.Items == nil {
			return nil
		}

		if err := deleteInvoiceItemsByInvoiceID(tx, invoice.ID); err != nil {
			return err
		}

		return saveItems(tx, invoice.ID, invoice.Items)
	})
}

func (ir *InvoiceRepository) CloseInvoice(id int, event model.OutboxEvent) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		if err := startTransition(tx, id, model.InvoiceStatusClosing); err != nil {
			return err
		}

		return tx.Create(&event).Error
	})
}

func (ir *InvoiceRepository) ConfirmClose(id int) (bool, error) {
	return ir.moveFrom(id, model.InvoiceStatusClosing, model.InvoiceStatusClosed, "")
}

func (ir *InvoiceRepository) RejectClose(id int, reason string) (bool, error) {
	return ir.moveFrom(id, model.InvoiceStatusClosing, model.InvoiceStatusOpen, reason)
}

func (ir *InvoiceRepository) ReopenInvoice(id int, event model.OutboxEvent) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		if err := startTransition(tx, id, model.InvoiceStatusReopening); err != nil {
			return err
		}

		return tx.Create(&event).Error
	})
}

func (ir *InvoiceRepository) ConfirmReopen(id int) (bool, error) {
	return ir.moveFrom(id, model.InvoiceStatusReopening, model.InvoiceStatusOpen, "")
}

func (ir *InvoiceRepository) RejectReopen(id int, reason string) (bool, error) {
	return ir.moveFrom(id, model.InvoiceStatusReopening, model.InvoiceStatusClosed, reason)
}

func (ir *InvoiceRepository) moveFrom(id int, from, to, reason string) (bool, error) {
	result := ir.connection.
		Model(&model.Invoice{}).
		Where("id = ? AND status = ?", id, from).
		Updates(map[string]any{"status": to, "failure_reason": reason})
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

func startTransition(tx *gorm.DB, id int, status string) error {
	result := tx.
		Model(&model.Invoice{ID: id}).
		Updates(map[string]any{"status": status, "failure_reason": ""})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (ir *InvoiceRepository) DeleteInvoice(id int) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		if err := deleteInvoiceItemsByInvoiceID(tx, id); err != nil {
			return err
		}

		result := tx.Delete(&model.Invoice{}, id)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func deleteInvoiceItemsByInvoiceID(tx *gorm.DB, invoiceID int) error {
	return tx.Where("invoice_id = ?", invoiceID).Delete(&model.InvoiceItem{}).Error
}

func withItems(connection *gorm.DB) *gorm.DB {
	return connection.
		Preload("Items", func(items *gorm.DB) *gorm.DB {
			return items.Order("invoice_item.id")
		}).
		Preload("Items.Product")
}

func saveItems(tx *gorm.DB, invoiceID int, items []model.InvoiceItem) error {
	for _, item := range items {
		product, err := ensureProduct(tx, item.Product)
		if err != nil {
			return err
		}

		item.ID = 0
		item.InvoiceID = invoiceID
		item.ProductID = product.ID

		if err := tx.Omit("Product").Create(&item).Error; err != nil {
			return err
		}
	}

	return nil
}

func ensureProduct(tx *gorm.DB, product model.Product) (model.Product, error) {
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "inventory_id"}},
		DoNothing: true,
	}).Create(&product).Error
	if err != nil {
		return model.Product{}, err
	}

	if product.ID != 0 {
		return product, nil
	}

	var stored model.Product
	if err := tx.Where("inventory_id = ?", product.InventoryID).First(&stored).Error; err != nil {
		return model.Product{}, err
	}

	return stored, nil
}

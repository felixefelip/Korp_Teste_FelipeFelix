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
			Select("number", "status").
			Updates(model.Invoice{Number: invoice.Number, Status: invoice.Status})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if invoice.Items == nil {
			return nil
		}

		err := tx.Where("invoice_id = ?", invoice.ID).Delete(&model.InvoiceItem{}).Error
		if err != nil {
			return err
		}

		return saveItems(tx, invoice.ID, invoice.Items)
	})
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
		product, err := saveProduct(tx, item.Product)
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

func saveProduct(tx *gorm.DB, product model.Product) (model.Product, error) {
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "inventory_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"code", "name", "unit", "price"}),
	}).Create(&product).Error
	if err != nil {
		return model.Product{}, err
	}

	return product, nil
}

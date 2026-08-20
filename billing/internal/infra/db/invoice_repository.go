package db

import (
	"billing/internal/model"

	"gorm.io/gorm"
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

	err := ir.connection.Find(&invoiceList).Error
	if err != nil {
		return []model.Invoice{}, err
	}

	return invoiceList, nil
}

func (ir *InvoiceRepository) GetInvoiceByID(id int) (model.Invoice, error) {
	var invoice model.Invoice

	err := ir.connection.First(&invoice, id).Error
	if err != nil {
		return model.Invoice{}, err
	}

	return invoice, nil
}

func (ir *InvoiceRepository) CreateInvoice(invoice model.Invoice) (int, error) {
	err := ir.connection.Create(&invoice).Error
	if err != nil {
		return 0, err
	}

	return invoice.ID, nil
}

func (ir *InvoiceRepository) UpdateInvoice(invoice model.Invoice) error {
	result := ir.connection.
		Model(&model.Invoice{ID: invoice.ID}).
		Select("number", "status").
		Updates(invoice)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

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

func (ir *InvoiceRepository) CreateInvoice(invoice model.Invoice) (int, error) {
	err := ir.connection.Create(&invoice).Error
	if err != nil {
		return 0, err
	}

	return invoice.ID, nil
}

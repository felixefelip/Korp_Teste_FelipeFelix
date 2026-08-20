package model

type InvoiceRepository interface {
	GetInvoices() ([]Invoice, error)
	GetInvoiceByID(id int) (Invoice, error)
	CreateInvoice(invoice Invoice) (int, error)
	UpdateInvoice(invoice Invoice) error
}

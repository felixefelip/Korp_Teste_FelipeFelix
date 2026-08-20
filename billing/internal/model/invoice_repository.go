package model

type InvoiceRepository interface {
	GetInvoices() ([]Invoice, error)
	CreateInvoice(invoice Invoice) (int, error)
}

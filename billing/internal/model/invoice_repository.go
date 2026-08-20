package model

type InvoiceRepository interface {
	CreateInvoice(invoice Invoice) (int, error)
}

package model

type InvoiceRepository interface {
	GetInvoices() ([]Invoice, error)
	GetInvoiceByID(id int) (Invoice, error)
	CreateInvoice(invoice Invoice) (int, error)
	UpdateInvoice(invoice Invoice) error
	CloseInvoice(id int, event OutboxEvent) error
	ConfirmClose(id int) (bool, error)
	RejectClose(id int, reason string) (bool, error)
	ReopenInvoice(id int) error
	DeleteInvoice(id int) error
}

package model

type InvoiceRepository interface {
	GetInvoices() ([]Invoice, error)
	GetInvoiceByID(id int) (Invoice, error)
	CreateInvoice(invoice Invoice) (int, error)
	UpdateInvoice(invoice Invoice) error
	CloseInvoice(id int, event OutboxEvent) error
	ConfirmClose(id int) (bool, error)
	RejectClose(id int, reason string, shortages []InvoiceShortage) (bool, error)
	ReopenInvoice(id int, event OutboxEvent) error
	RetryInvoice(id int, status string, event OutboxEvent) error
	ConfirmReopen(id int) (bool, error)
	RejectReopen(id int, reason string, shortages []InvoiceShortage) (bool, error)
	DeleteInvoice(id int) error
}

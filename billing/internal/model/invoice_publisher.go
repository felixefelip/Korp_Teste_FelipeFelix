package model

type InvoiceEventPublisher interface {
	PublishCloseRequested(invoice Invoice) error
}

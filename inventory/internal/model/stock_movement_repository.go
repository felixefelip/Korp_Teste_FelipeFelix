package model

type StockMovementRepository interface {
	GetMovementsByProductID(productID int) ([]StockMovement, error)
	GetMovementByID(id int) (StockMovement, error)
	CreateMovement(movement StockMovement) (int, error)
	UpdateMovement(movement StockMovement) error
	ApplyInvoice(request InvoiceStockRequest, event OutboxEvent) error
}

package messaging

import (
	"inventory/internal/infra/db"
	"inventory/internal/infra/messaging/invoice"
	"inventory/internal/model"
	"inventory/internal/usecase"

	"gorm.io/gorm"
)

func Register(connection *gorm.DB) {
	movementRepository := db.NewStockMovementRepository(connection)
	outboxRepository := db.NewOutboxRepository(connection)

	invoiceStockUsecase := usecase.NewInvoiceStockUsecase(movementRepository)
	invoiceHandler := invoice.NewHandler(invoiceStockUsecase)

	NewConsumer(InvoiceRequestsQueue, Routes{
		model.InvoiceCloseRequestedKey: invoiceHandler.HandleCloseRequested,
	}).Start()

	NewRelay(outboxRepository, InventoryExchange).Start()
}

package messaging

import (
	"billing/internal/infra/db"
	"billing/internal/infra/messaging/invoice"
	"billing/internal/infra/messaging/product"
	"billing/internal/model"
	"billing/internal/usecase"

	"gorm.io/gorm"
)

func Register(connection *gorm.DB) {
	invoiceRepository := db.NewInvoiceRepository(connection)
	outboxRepository := db.NewOutboxRepository(connection)

	invoiceUsecase := usecase.NewInvoiceUsecase(invoiceRepository)
	invoiceHandler := invoice.NewHandler(invoiceUsecase)

	NewConsumer(StockResultsQueue, Routes{
		model.InvoiceStockAppliedKey:  invoiceHandler.HandleStockApplied,
		model.InvoiceStockRejectedKey: invoiceHandler.HandleStockRejected,

		model.InvoiceStockRevertedKey:       invoiceHandler.HandleStockReverted,
		model.InvoiceStockRevertRejectedKey: invoiceHandler.HandleStockRevertRejected,
	}).Start()

	productUsecase := usecase.NewProductUsecase(db.NewProductRepository(connection))
	productHandler := product.NewHandler(productUsecase)

	NewConsumer(CatalogQueue, Routes{
		model.ProductCreatedKey: productHandler.HandleProductSaved,
		model.ProductUpdatedKey: productHandler.HandleProductSaved,
		model.ProductDeletedKey: productHandler.HandleProductDeleted,
	}).Start()

	NewRelay(outboxRepository, BillingExchange).Start()
}

package web

import (
	"net/http"

	"billing/internal/infra/db"
	"billing/internal/infra/web/invoice"
	"billing/internal/model"
	"billing/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(connection *gorm.DB, publisher model.InvoiceEventPublisher) *gin.Engine {
	server := gin.Default()

	Register(server, connection, publisher)

	return server
}

func Register(server *gin.Engine, connection *gorm.DB, publisher model.InvoiceEventPublisher) {
	invoiceRepository := db.NewInvoiceRepository(connection)
	invoiceUsecase := usecase.NewInvoiceUsecase(invoiceRepository, publisher)
	invoiceController := invoice.NewController(invoiceUsecase)

	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	server.GET("/invoices", invoiceController.GetInvoices)
	server.GET("/invoices/:id", invoiceController.GetInvoiceByID)
	server.POST("/invoices", invoiceController.CreateInvoice)
	server.PUT("/invoices/:id", invoiceController.UpdateInvoice)
	server.POST("/invoices/:id/close", invoiceController.CloseInvoice)
	server.POST("/invoices/:id/reopen", invoiceController.ReopenInvoice)
	server.DELETE("/invoices/:id", invoiceController.DeleteInvoice)
}

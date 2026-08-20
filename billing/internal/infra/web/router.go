package web

import (
	"net/http"

	"billing/internal/infra/db"
	"billing/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(connection *gorm.DB) *gin.Engine {
	server := gin.Default()

	Register(server, connection)

	return server
}

func Register(server *gin.Engine, connection *gorm.DB) {
	invoiceRepository := db.NewInvoiceRepository(connection)
	invoiceUsecase := usecase.NewInvoiceUsecase(invoiceRepository)
	invoiceController := NewInvoiceController(invoiceUsecase)

	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	server.GET("/invoices", invoiceController.GetInvoices)
	server.GET("/invoices/:id", invoiceController.GetInvoiceByID)
	server.POST("/invoices", invoiceController.CreateInvoice)
	server.PUT("/invoices/:id", invoiceController.UpdateInvoice)
}

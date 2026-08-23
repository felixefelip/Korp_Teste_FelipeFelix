package web

import (
	"net/http"

	"billing/internal/infra/db"
	"billing/internal/infra/web/invoice"
	"billing/internal/infra/web/product"
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
	invoiceController := invoice.NewController(invoiceUsecase)

	productRepository := db.NewProductRepository(connection)
	productUsecase := usecase.NewProductUsecase(productRepository)
	productController := product.NewController(productUsecase)

	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	server.GET("/products", productController.GetProducts)

	server.GET("/invoices", invoiceController.GetInvoices)
	server.GET("/invoices/:id", invoiceController.GetInvoiceByID)
	server.POST("/invoices", invoiceController.CreateInvoice)
	server.PUT("/invoices/:id", invoiceController.UpdateInvoice)
	server.POST("/invoices/:id/close", invoiceController.CloseInvoice)
	server.POST("/invoices/:id/reopen", invoiceController.ReopenInvoice)
	server.POST("/invoices/:id/retry", invoiceController.RetryInvoice)
	server.GET("/invoices/:id/danfe", invoiceController.PrintInvoice)
	server.DELETE("/invoices/:id", invoiceController.DeleteInvoice)
}

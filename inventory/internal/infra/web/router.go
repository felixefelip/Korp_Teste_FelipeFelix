package web

import (
	"net/http"

	"inventory/internal/infra/db"
	"inventory/internal/infra/web/product"
	"inventory/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(connection *gorm.DB) *gin.Engine {
	server := gin.Default()

	Register(server, connection)

	return server
}

func Register(server *gin.Engine, connection *gorm.DB) {
	productRepository := db.NewProductRepository(connection)
	productUsecase := usecase.NewProductUsecase(productRepository)
	productController := product.NewController(productUsecase)

	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	server.GET("/products", productController.GetProducts)
	server.GET("/products/:id", productController.GetProductByID)
	server.POST("/products", productController.CreateProduct)
	server.PUT("/products/:id", productController.UpdateProduct)
}

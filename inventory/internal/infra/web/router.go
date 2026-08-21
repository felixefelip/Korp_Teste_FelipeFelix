package web

import (
	"net/http"

	"inventory/internal/infra/db"
	"inventory/internal/infra/web/movement"
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
	movementRepository := db.NewStockMovementRepository(connection)

	productUsecase := usecase.NewProductUsecase(productRepository, movementRepository)
	productController := product.NewController(productUsecase)

	movementUsecase := usecase.NewStockMovementUsecase(movementRepository, productRepository)
	movementController := movement.NewController(movementUsecase)

	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	server.GET("/products", productController.GetProducts)
	server.GET("/products/:id", productController.GetProductByID)
	server.POST("/products", productController.CreateProduct)
	server.PUT("/products/:id", productController.UpdateProduct)
	server.DELETE("/products/:id", productController.DeleteProduct)

	server.GET("/products/:id/movements", movementController.GetMovements)
	server.POST("/products/:id/movements", movementController.CreateMovement)
	server.GET("/products/:id/movements/:movementId", movementController.GetMovementByID)
	server.PUT("/products/:id/movements/:movementId", movementController.UpdateMovement)
}

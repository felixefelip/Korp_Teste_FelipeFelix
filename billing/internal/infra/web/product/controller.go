package product

import (
	"net/http"

	"billing/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	productUsecase usecase.ProductUsecase
}

func NewController(usecase usecase.ProductUsecase) Controller {
	return Controller{
		productUsecase: usecase,
	}
}

func (p *Controller) GetProducts(ctx *gin.Context) {
	products, err := p.productUsecase.GetProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao buscar os produtos"})
		return
	}

	ctx.JSON(http.StatusOK, newResponses(products))
}

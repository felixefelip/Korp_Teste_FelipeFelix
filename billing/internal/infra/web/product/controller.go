package product

import (
	"net/http"

	"billing/internal/infra/web/apierr"
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
		apierr.Internal(ctx, "erro ao buscar os produtos", err)
		return
	}

	ctx.JSON(http.StatusOK, newResponses(products))
}

package product

import (
	"errors"
	"net/http"
	"strconv"

	"inventory/internal/infra/web/apierr"
	"inventory/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, newResponses(products))
}

func (p *Controller) GetProductByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id do produto precisa ser um numero inteiro"})
		return
	}

	product, err := p.productUsecase.GetProductByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "produto nao encontrado"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao buscar o produto"})
		return
	}

	ctx.JSON(http.StatusOK, newResponse(product))
}

func (p *Controller) UpdateProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id do produto precisa ser um numero inteiro"})
		return
	}

	var request updateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		if fieldErrors := apierr.FieldErrors(err); fieldErrors != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"errors": fieldErrors})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Não foi possível ler os dados enviados."})
		return
	}

	updatedProduct, err := p.productUsecase.UpdateProduct(request.toModel(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "produto nao encontrado"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao atualizar o produto"})
		return
	}

	ctx.JSON(http.StatusOK, newResponse(updatedProduct))
}

func (p *Controller) CreateProduct(ctx *gin.Context) {
	var request createRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		if fieldErrors := apierr.FieldErrors(err); fieldErrors != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"errors": fieldErrors})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Não foi possível ler os dados enviados."})
		return
	}

	insertedProduct, err := p.productUsecase.CreateProduct(request.toModel())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao criar o produto"})
		return
	}

	ctx.JSON(http.StatusCreated, newResponse(insertedProduct))
}

func (p *Controller) DeleteProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id do produto precisa ser um numero inteiro"})
		return
	}

	if err := p.productUsecase.DeleteProduct(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "produto nao encontrado"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao excluir o produto"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

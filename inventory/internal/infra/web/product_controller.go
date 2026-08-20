package web

import (
	"errors"
	"net/http"
	"strconv"

	"inventory/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type productController struct {
	productUsecase usecase.ProductUsecase
}

func NewProductController(usecase usecase.ProductUsecase) productController {
	return productController{
		productUsecase: usecase,
	}
}

func (p *productController) GetProducts(ctx *gin.Context) {
	products, err := p.productUsecase.GetProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, newProductResponses(products))
}

func (p *productController) GetProductByID(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, newProductResponse(product))
}

func (p *productController) UpdateProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id do produto precisa ser um numero inteiro"})
		return
	}

	var request updateProductRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		if fieldErrors := bindErrors(err); fieldErrors != nil {
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

	ctx.JSON(http.StatusOK, newProductResponse(updatedProduct))
}

func (p *productController) CreateProduct(ctx *gin.Context) {
	var request createProductRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		if fieldErrors := bindErrors(err); fieldErrors != nil {
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

	ctx.JSON(http.StatusCreated, newProductResponse(insertedProduct))
}

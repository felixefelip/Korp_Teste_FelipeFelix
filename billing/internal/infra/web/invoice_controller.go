package web

import (
	"errors"
	"net/http"
	"strconv"

	"billing/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type invoiceController struct {
	invoiceUsecase usecase.InvoiceUsecase
}

func NewInvoiceController(usecase usecase.InvoiceUsecase) invoiceController {
	return invoiceController{
		invoiceUsecase: usecase,
	}
}

func (i *invoiceController) GetInvoices(ctx *gin.Context) {
	invoices, err := i.invoiceUsecase.GetInvoices()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao buscar as notas fiscais"})
		return
	}

	ctx.JSON(http.StatusOK, newInvoiceResponses(invoices))
}

func (i *invoiceController) GetInvoiceByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id da nota fiscal precisa ser um numero inteiro"})
		return
	}

	invoice, err := i.invoiceUsecase.GetInvoiceByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao buscar a nota fiscal"})
		return
	}

	ctx.JSON(http.StatusOK, newInvoiceResponse(invoice))
}

func (i *invoiceController) UpdateInvoice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id da nota fiscal precisa ser um numero inteiro"})
		return
	}

	var request updateInvoiceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		if fieldErrors := bindErrors(err); fieldErrors != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"errors": fieldErrors})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Não foi possível ler os dados enviados."})
		return
	}

	updatedInvoice, err := i.invoiceUsecase.UpdateInvoice(request.toModel(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao atualizar a nota fiscal"})
		return
	}

	ctx.JSON(http.StatusOK, newInvoiceResponse(updatedInvoice))
}

func (i *invoiceController) CreateInvoice(ctx *gin.Context) {
	var request createInvoiceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		if fieldErrors := bindErrors(err); fieldErrors != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"errors": fieldErrors})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Não foi possível ler os dados enviados."})
		return
	}

	insertedInvoice, err := i.invoiceUsecase.CreateInvoice(request.toModel())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao criar a nota fiscal"})
		return
	}

	ctx.JSON(http.StatusCreated, newInvoiceResponse(insertedInvoice))
}

package invoice

import (
	"errors"
	"net/http"
	"strconv"

	"billing/internal/infra/web/apierr"
	"billing/internal/model"
	"billing/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	invoiceUsecase usecase.InvoiceUsecase
}

func NewController(usecase usecase.InvoiceUsecase) Controller {
	return Controller{
		invoiceUsecase: usecase,
	}
}

func (i *Controller) GetInvoices(ctx *gin.Context) {
	invoices, err := i.invoiceUsecase.GetInvoices()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao buscar as notas fiscais"})
		return
	}

	ctx.JSON(http.StatusOK, newResponses(invoices))
}

func (i *Controller) GetInvoiceByID(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, newResponse(invoice))
}

func (i *Controller) UpdateInvoice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id da nota fiscal precisa ser um numero inteiro"})
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

	updatedInvoice, err := i.invoiceUsecase.UpdateInvoice(request.toModel(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao atualizar a nota fiscal"})
		return
	}

	ctx.JSON(http.StatusOK, newResponse(updatedInvoice))
}

func (i *Controller) CreateInvoice(ctx *gin.Context) {
	var request createRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		if fieldErrors := apierr.FieldErrors(err); fieldErrors != nil {
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

	ctx.JSON(http.StatusCreated, newResponse(insertedInvoice))
}

func (i *Controller) DeleteInvoice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id da nota fiscal precisa ser um numero inteiro"})
		return
	}

	if err := i.invoiceUsecase.DeleteInvoice(id); err != nil {
		if errors.Is(err, model.ErrInvoiceClosed) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Notas fiscais fechadas não podem ser excluídas.",
			})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao excluir a nota fiscal"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

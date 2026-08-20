package web

import (
	"net/http"

	"billing/internal/usecase"

	"github.com/gin-gonic/gin"
)

type invoiceController struct {
	invoiceUsecase usecase.InvoiceUsecase
}

func NewInvoiceController(usecase usecase.InvoiceUsecase) invoiceController {
	return invoiceController{
		invoiceUsecase: usecase,
	}
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

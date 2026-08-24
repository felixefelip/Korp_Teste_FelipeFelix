package invoice

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"billing/internal/infra/pdf"
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
		apierr.Internal(ctx, "erro ao buscar as notas fiscais", err)
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

		apierr.Internal(ctx, "erro ao buscar a nota fiscal", err)
		return
	}

	ctx.JSON(http.StatusOK, newResponse(invoice))
}

func (i *Controller) GetNextDocument(ctx *gin.Context) {
	series := 0

	if raw := ctx.Query("series"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "a série precisa ser um numero inteiro positivo",
			})
			return
		}

		series = parsed
	}

	document, err := i.invoiceUsecase.NextDocument(series)
	if err != nil {
		apierr.Internal(ctx, "erro ao sugerir o numero da nota fiscal", err)
		return
	}

	ctx.JSON(http.StatusOK, newDocumentResponse(document))
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
		if errors.Is(err, model.ErrInvoiceDuplicated) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Já existe uma nota fiscal com esta série e número.",
			})
			return
		}

		if errors.Is(err, model.ErrInvoiceProcessing) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Esta nota fiscal está em processamento.",
			})
			return
		}

		if errors.Is(err, model.ErrInvoiceClosed) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Notas fiscais fechadas não podem ser alteradas.",
			})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		apierr.Internal(ctx, "erro ao atualizar a nota fiscal", err)
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
		if errors.Is(err, model.ErrInvoiceDuplicated) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Já existe uma nota fiscal com esta série e número.",
			})
			return
		}

		apierr.Internal(ctx, "erro ao criar a nota fiscal", err)
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
		if errors.Is(err, model.ErrInvoiceProcessing) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Esta nota fiscal está em processamento.",
			})
			return
		}

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

		apierr.Internal(ctx, "erro ao excluir a nota fiscal", err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (i *Controller) CloseInvoice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id da nota fiscal precisa ser um numero inteiro"})
		return
	}

	closedInvoice, err := i.invoiceUsecase.CloseInvoice(id)
	if err != nil {
		if errors.Is(err, model.ErrInvoiceProcessing) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Esta nota fiscal está em processamento.",
			})
			return
		}

		if errors.Is(err, model.ErrInvoiceClosed) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Esta nota fiscal já está fechada.",
			})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		apierr.Internal(ctx, "erro ao fechar a nota fiscal", err)
		return
	}

	ctx.JSON(http.StatusAccepted, newResponse(closedInvoice))
}

func (i *Controller) PrintInvoice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id da nota fiscal precisa ser um numero inteiro"})
		return
	}

	invoice, err := i.invoiceUsecase.GetInvoiceToPrint(id)
	if err != nil {
		if errors.Is(err, model.ErrInvoiceProcessing) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Esta nota fiscal está em processamento.",
			})
			return
		}

		if errors.Is(err, model.ErrInvoiceOpen) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "A DANFE só pode ser impressa depois que a nota fiscal for fechada.",
			})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		apierr.Internal(ctx, "erro ao buscar a nota fiscal", err)
		return
	}

	document, err := pdf.Danfe(invoice)
	if err != nil {
		apierr.Internal(ctx, "erro ao gerar a DANFE", err)
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf(`inline; filename="danfe-%s.pdf"`, filename(invoice)))
	ctx.Data(http.StatusOK, "application/pdf", document)
}

func filename(invoice model.Invoice) string {
	return strings.ReplaceAll(invoice.FormattedNumber(), "/", "-")
}

func (i *Controller) ReopenInvoice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id da nota fiscal precisa ser um numero inteiro"})
		return
	}

	reopenedInvoice, err := i.invoiceUsecase.ReopenInvoice(id)
	if err != nil {
		if errors.Is(err, model.ErrInvoiceProcessing) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Esta nota fiscal está em processamento.",
			})
			return
		}

		if errors.Is(err, model.ErrInvoiceOpen) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Esta nota fiscal já está aberta.",
			})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		apierr.Internal(ctx, "erro ao reabrir a nota fiscal", err)
		return
	}

	ctx.JSON(http.StatusAccepted, newResponse(reopenedInvoice))
}

func (i *Controller) RetryInvoice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "o id da nota fiscal precisa ser um numero inteiro"})
		return
	}

	retriedInvoice, err := i.invoiceUsecase.RetryInvoice(id)
	if err != nil {
		if errors.Is(err, model.ErrInvoiceNotProcessing) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Esta nota fiscal não está em processamento.",
			})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "nota fiscal nao encontrada"})
			return
		}

		apierr.Internal(ctx, "erro ao reenviar a nota fiscal", err)
		return
	}

	ctx.JSON(http.StatusAccepted, newResponse(retriedInvoice))
}

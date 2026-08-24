package invoice

import (
	"errors"
	"net/http"

	"billing/internal/infra/web/apierr"
	"billing/internal/model"
	"billing/internal/usecase"

	"github.com/gin-gonic/gin"
)

type DraftController struct {
	draftUsecase usecase.InvoiceDraftUsecase
}

func NewDraftController(usecase usecase.InvoiceDraftUsecase) DraftController {
	return DraftController{
		draftUsecase: usecase,
	}
}

func (d *DraftController) DraftInvoice(ctx *gin.Context) {
	var request draftRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		if fieldErrors := apierr.FieldErrors(err); fieldErrors != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"errors": fieldErrors})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Não foi possível ler os dados enviados."})
		return
	}

	draft, err := d.draftUsecase.DraftInvoice(ctx.Request.Context(), request.Prompt)
	if err != nil {
		if errors.Is(err, model.ErrDraftUnavailable) {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "O preenchimento por IA não está configurado neste ambiente.",
			})
			return
		}

		apierr.Log(ctx, err)

		ctx.JSON(http.StatusBadGateway, gin.H{
			"message": "Não foi possível interpretar o pedido agora. Tente novamente.",
		})
		return
	}

	ctx.JSON(http.StatusOK, newDraftResponse(draft))
}

package movement

import (
	"errors"
	"net/http"
	"strconv"

	"inventory/internal/infra/web/apierr"
	"inventory/internal/model"
	"inventory/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	movementUsecase usecase.StockMovementUsecase
}

func NewController(usecase usecase.StockMovementUsecase) Controller {
	return Controller{
		movementUsecase: usecase,
	}
}

func (m *Controller) GetMovements(ctx *gin.Context) {
	productID, ok := param(ctx, "id", "o id do produto precisa ser um numero inteiro")
	if !ok {
		return
	}

	movements, err := m.movementUsecase.GetMovementsByProductID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "produto nao encontrado"})
			return
		}

		apierr.Internal(ctx, "erro ao buscar as movimentacoes", err)
		return
	}

	ctx.JSON(http.StatusOK, newResponses(movements))
}

func (m *Controller) GetMovementByID(ctx *gin.Context) {
	productID, ok := param(ctx, "id", "o id do produto precisa ser um numero inteiro")
	if !ok {
		return
	}

	movementID, ok := param(ctx, "movementId", "o id da movimentacao precisa ser um numero inteiro")
	if !ok {
		return
	}

	movement, err := m.movementUsecase.GetMovementByID(movementID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(ctx)
			return
		}

		apierr.Internal(ctx, "erro ao buscar a movimentacao", err)
		return
	}

	if movement.ProductID != productID {
		notFound(ctx)
		return
	}

	ctx.JSON(http.StatusOK, newResponse(movement))
}

func (m *Controller) CreateMovement(ctx *gin.Context) {
	productID, ok := param(ctx, "id", "o id do produto precisa ser um numero inteiro")
	if !ok {
		return
	}

	var request createRequest
	if !bind(ctx, &request) {
		return
	}

	createdMovement, err := m.movementUsecase.CreateMovement(request.toModel(productID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "produto nao encontrado"})
			return
		}

		apierr.Internal(ctx, "erro ao criar a movimentacao", err)
		return
	}

	ctx.JSON(http.StatusCreated, newResponse(createdMovement))
}

func (m *Controller) UpdateMovement(ctx *gin.Context) {
	productID, ok := param(ctx, "id", "o id do produto precisa ser um numero inteiro")
	if !ok {
		return
	}

	movementID, ok := param(ctx, "movementId", "o id da movimentacao precisa ser um numero inteiro")
	if !ok {
		return
	}

	stored, err := m.movementUsecase.GetMovementByID(movementID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(ctx)
			return
		}

		apierr.Internal(ctx, "erro ao buscar a movimentacao", err)
		return
	}

	if stored.ProductID != productID {
		notFound(ctx)
		return
	}

	var request updateRequest
	if !bind(ctx, &request) {
		return
	}

	updatedMovement, err := m.movementUsecase.UpdateMovement(request.toModel(movementID))
	if err != nil {
		if errors.Is(err, model.ErrMovementFromInvoice) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "Movimentações geradas por notas fiscais não podem ser alteradas.",
			})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(ctx)
			return
		}

		apierr.Internal(ctx, "erro ao atualizar a movimentacao", err)
		return
	}

	ctx.JSON(http.StatusOK, newResponse(updatedMovement))
}

func param(ctx *gin.Context, name, message string) (int, bool) {
	value, err := strconv.Atoi(ctx.Param(name))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": message})
		return 0, false
	}

	return value, true
}

func bind(ctx *gin.Context, request any) bool {
	if err := ctx.ShouldBindJSON(request); err != nil {
		if fieldErrors := apierr.FieldErrors(err); fieldErrors != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"errors": fieldErrors})
			return false
		}

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Não foi possível ler os dados enviados."})
		return false
	}

	return true
}

func notFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, gin.H{"message": "movimentacao nao encontrada"})
}

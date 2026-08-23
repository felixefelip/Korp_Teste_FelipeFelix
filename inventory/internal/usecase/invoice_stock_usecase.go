package usecase

import (
	"inventory/internal/model"
)

type InvoiceStockUsecase struct {
	repository model.StockMovementRepository
}

func NewInvoiceStockUsecase(repository model.StockMovementRepository) InvoiceStockUsecase {
	return InvoiceStockUsecase{
		repository: repository,
	}
}

func (iu *InvoiceStockUsecase) Apply(request model.InvoiceStockRequest) (model.OutboxEvent, error) {
	return iu.repository.ApplyInvoice(request)
}

func (iu *InvoiceStockUsecase) Revert(
	request model.InvoiceStockRevertRequest,
) (model.OutboxEvent, error) {
	return iu.repository.RevertInvoice(request)
}

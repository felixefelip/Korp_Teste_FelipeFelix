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

func (iu *InvoiceStockUsecase) Apply(request model.InvoiceStockRequest) error {
	event, err := model.NewInvoiceStockApplied(request.InvoiceID)
	if err != nil {
		return err
	}

	return iu.repository.ApplyInvoice(request, event)
}

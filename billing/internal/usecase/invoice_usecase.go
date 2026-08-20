package usecase

import (
	"billing/internal/model"
)

type InvoiceUsecase struct {
	repository model.InvoiceRepository
}

func NewInvoiceUsecase(repository model.InvoiceRepository) InvoiceUsecase {
	return InvoiceUsecase{
		repository: repository,
	}
}

func (iu *InvoiceUsecase) CreateInvoice(invoice model.Invoice) (model.Invoice, error) {
	invoiceId, err := iu.repository.CreateInvoice(invoice)
	if err != nil {
		return model.Invoice{}, err
	}

	invoice.ID = invoiceId

	return invoice, nil
}

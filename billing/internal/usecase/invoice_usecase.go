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

func (iu *InvoiceUsecase) GetInvoices() ([]model.Invoice, error) {
	return iu.repository.GetInvoices()
}

func (iu *InvoiceUsecase) GetInvoiceByID(id int) (model.Invoice, error) {
	return iu.repository.GetInvoiceByID(id)
}

func (iu *InvoiceUsecase) UpdateInvoice(invoice model.Invoice) (model.Invoice, error) {
	if err := iu.repository.UpdateInvoice(invoice); err != nil {
		return model.Invoice{}, err
	}

	return iu.repository.GetInvoiceByID(invoice.ID)
}

func (iu *InvoiceUsecase) CreateInvoice(invoice model.Invoice) (model.Invoice, error) {
	invoiceId, err := iu.repository.CreateInvoice(invoice)
	if err != nil {
		return model.Invoice{}, err
	}

	return iu.repository.GetInvoiceByID(invoiceId)
}

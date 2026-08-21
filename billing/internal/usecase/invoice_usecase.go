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

func (iu *InvoiceUsecase) CloseInvoice(id int) (model.Invoice, error) {
	invoice, err := iu.repository.GetInvoiceByID(id)
	if err != nil {
		return model.Invoice{}, err
	}

	if invoice.Closed() {
		return model.Invoice{}, model.ErrInvoiceClosed
	}

	if err := iu.repository.CloseInvoice(id); err != nil {
		return model.Invoice{}, err
	}

	return iu.repository.GetInvoiceByID(id)
}

func (iu *InvoiceUsecase) ReopenInvoice(id int) (model.Invoice, error) {
	invoice, err := iu.repository.GetInvoiceByID(id)
	if err != nil {
		return model.Invoice{}, err
	}

	if !invoice.Closed() {
		return model.Invoice{}, model.ErrInvoiceOpen
	}

	if err := iu.repository.ReopenInvoice(id); err != nil {
		return model.Invoice{}, err
	}

	return iu.repository.GetInvoiceByID(id)
}

func (iu *InvoiceUsecase) DeleteInvoice(id int) error {
	invoice, err := iu.repository.GetInvoiceByID(id)
	if err != nil {
		return err
	}

	if invoice.Closed() {
		return model.ErrInvoiceClosed
	}

	return iu.repository.DeleteInvoice(id)
}

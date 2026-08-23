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
	stored, err := iu.repository.GetInvoiceByID(invoice.ID)
	if err != nil {
		return model.Invoice{}, err
	}

	if !stored.Editable() {
		return model.Invoice{}, blockedByStatus(stored)
	}

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

	if !invoice.Editable() {
		return model.Invoice{}, blockedByStatus(invoice)
	}

	event, err := model.NewInvoiceCloseRequested(invoice)
	if err != nil {
		return model.Invoice{}, err
	}

	if err := iu.repository.CloseInvoice(id, event); err != nil {
		return model.Invoice{}, err
	}

	return iu.repository.GetInvoiceByID(id)
}

func (iu *InvoiceUsecase) ReopenInvoice(id int) (model.Invoice, error) {
	invoice, err := iu.repository.GetInvoiceByID(id)
	if err != nil {
		return model.Invoice{}, err
	}

	if invoice.Processing() {
		return model.Invoice{}, model.ErrInvoiceProcessing
	}

	if !invoice.Closed() {
		return model.Invoice{}, model.ErrInvoiceOpen
	}

	event, err := model.NewInvoiceReopenRequested(invoice)
	if err != nil {
		return model.Invoice{}, err
	}

	if err := iu.repository.ReopenInvoice(id, event); err != nil {
		return model.Invoice{}, err
	}

	return iu.repository.GetInvoiceByID(id)
}

func (iu *InvoiceUsecase) GetInvoiceToPrint(id int) (model.Invoice, error) {
	invoice, err := iu.repository.GetInvoiceByID(id)
	if err != nil {
		return model.Invoice{}, err
	}

	if invoice.Processing() {
		return model.Invoice{}, model.ErrInvoiceProcessing
	}

	if !invoice.Closed() {
		return model.Invoice{}, model.ErrInvoiceOpen
	}

	return invoice, nil
}

func (iu *InvoiceUsecase) DeleteInvoice(id int) error {
	invoice, err := iu.repository.GetInvoiceByID(id)
	if err != nil {
		return err
	}

	if !invoice.Editable() {
		return blockedByStatus(invoice)
	}

	return iu.repository.DeleteInvoice(id)
}

func (iu *InvoiceUsecase) ConfirmClose(invoiceID int) error {
	applied, err := iu.repository.ConfirmClose(invoiceID)
	if err != nil {
		return err
	}

	if !applied {
		return model.ErrInvoiceNotProcessing
	}

	return nil
}

func (iu *InvoiceUsecase) RejectClose(
	invoiceID int,
	reason string,
	shortages []model.InvoiceShortage,
) error {
	applied, err := iu.repository.RejectClose(invoiceID, reason, shortages)
	if err != nil {
		return err
	}

	if !applied {
		return model.ErrInvoiceNotProcessing
	}

	return nil
}

func (iu *InvoiceUsecase) ConfirmReopen(invoiceID int) error {
	applied, err := iu.repository.ConfirmReopen(invoiceID)
	if err != nil {
		return err
	}

	if !applied {
		return model.ErrInvoiceNotProcessing
	}

	return nil
}

func (iu *InvoiceUsecase) RejectReopen(
	invoiceID int,
	reason string,
	shortages []model.InvoiceShortage,
) error {
	applied, err := iu.repository.RejectReopen(invoiceID, reason, shortages)
	if err != nil {
		return err
	}

	if !applied {
		return model.ErrInvoiceNotProcessing
	}

	return nil
}

func blockedByStatus(invoice model.Invoice) error {
	if invoice.Processing() {
		return model.ErrInvoiceProcessing
	}

	return model.ErrInvoiceClosed
}

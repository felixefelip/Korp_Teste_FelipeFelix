package web

import (
	"strings"

	"billing/internal/model"
)

type createInvoiceRequest struct {
	Number string `json:"number" binding:"required,max=30"`
	Status string `json:"status" binding:"required,oneof=OPEN CLOSED"`
}

func (r createInvoiceRequest) toModel() model.Invoice {
	return model.Invoice{
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Status: r.Status,
	}
}

type updateInvoiceRequest struct {
	Number string `json:"number" binding:"required,max=30"`
	Status string `json:"status" binding:"required,oneof=OPEN CLOSED"`
}

func (r updateInvoiceRequest) toModel(id int) model.Invoice {
	return model.Invoice{
		ID:     id,
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Status: r.Status,
	}
}

type invoiceResponse struct {
	ID     int    `json:"id"`
	Number string `json:"number"`
	Status string `json:"status"`
}

func newInvoiceResponse(invoice model.Invoice) invoiceResponse {
	return invoiceResponse{
		ID:     invoice.ID,
		Number: invoice.Number,
		Status: invoice.Status,
	}
}

func newInvoiceResponses(invoices []model.Invoice) []invoiceResponse {
	responses := make([]invoiceResponse, 0, len(invoices))

	for _, invoice := range invoices {
		responses = append(responses, newInvoiceResponse(invoice))
	}

	return responses
}

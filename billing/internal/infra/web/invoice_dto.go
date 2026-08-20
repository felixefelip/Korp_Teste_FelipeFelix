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

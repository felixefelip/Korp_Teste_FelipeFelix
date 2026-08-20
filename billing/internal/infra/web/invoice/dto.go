package invoice

import (
	"math"
	"strings"

	"billing/internal/model"
)

type createRequest struct {
	Number string        `json:"number" binding:"required,max=30"`
	Status string        `json:"status" binding:"required,oneof=OPEN CLOSED"`
	Items  []itemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r createRequest) toModel() model.Invoice {
	return model.Invoice{
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Status: r.Status,
		Items:  toItemModels(r.Items),
	}
}

type updateRequest struct {
	Number string        `json:"number" binding:"required,max=30"`
	Status string        `json:"status" binding:"required,oneof=OPEN CLOSED"`
	Items  []itemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r updateRequest) toModel(id int) model.Invoice {
	return model.Invoice{
		ID:     id,
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Status: r.Status,
		Items:  toItemModels(r.Items),
	}
}

type response struct {
	ID     int            `json:"id"`
	Number string         `json:"number"`
	Status string         `json:"status"`
	Items  []itemResponse `json:"items"`
	Total  float64        `json:"total"`
}

func newResponse(invoice model.Invoice) response {
	items := make([]itemResponse, 0, len(invoice.Items))

	for _, item := range invoice.Items {
		items = append(items, newItemResponse(item))
	}

	return response{
		ID:     invoice.ID,
		Number: invoice.Number,
		Status: invoice.Status,
		Items:  items,
		Total:  round(invoice.Total()),
	}
}

func newResponses(invoices []model.Invoice) []response {
	responses := make([]response, 0, len(invoices))

	for _, invoice := range invoices {
		responses = append(responses, newResponse(invoice))
	}

	return responses
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

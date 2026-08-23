package invoice

import (
	"math"

	"billing/internal/model"
)

type response struct {
	ID     int            `json:"id"`
	Number string         `json:"number"`
	Type   string         `json:"type"`
	Status string         `json:"status"`
	Items  []itemResponse `json:"items"`
	Total  float64        `json:"total"`

	FailureReason string             `json:"failureReason,omitempty"`
	Shortages     []shortageResponse `json:"shortages,omitempty"`
}

type shortageResponse struct {
	InventoryID int    `json:"inventoryId"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Required    int    `json:"required"`
	Available   int    `json:"available"`
}

func newShortageResponses(shortages []model.InvoiceShortage) []shortageResponse {
	if len(shortages) == 0 {
		return nil
	}

	responses := make([]shortageResponse, 0, len(shortages))

	for _, shortage := range shortages {
		responses = append(responses, shortageResponse{
			InventoryID: shortage.InventoryID,
			Code:        shortage.ProductCode,
			Name:        shortage.ProductName,
			Required:    shortage.Required,
			Available:   shortage.Available,
		})
	}

	return responses
}

func newResponse(invoice model.Invoice) response {
	items := make([]itemResponse, 0, len(invoice.Items))

	for _, item := range invoice.Items {
		items = append(items, newItemResponse(item))
	}

	return response{
		ID:     invoice.ID,
		Number: invoice.Number,
		Type:   invoice.Type,
		Status: invoice.Status,
		Items:  items,
		Total:  round(invoice.Total()),

		FailureReason: invoice.FailureReason,
		Shortages:     newShortageResponses(invoice.Shortages),
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

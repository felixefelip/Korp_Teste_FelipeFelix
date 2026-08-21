package movement

import (
	"inventory/internal/model"
)

type response struct {
	ID            int    `json:"id"`
	ProductID     int    `json:"productId"`
	Type          string `json:"type"`
	Origin        string `json:"origin"`
	Quantity      int    `json:"quantity"`
	Confirmed     bool   `json:"confirmed"`
	InvoiceItemID *int   `json:"invoiceItemId"`
}

func newResponse(movement model.StockMovement) response {
	return response{
		ID:            movement.ID,
		ProductID:     movement.ProductID,
		Type:          movement.Type,
		Origin:        movement.Origin,
		Quantity:      movement.Quantity,
		Confirmed:     movement.Confirmed,
		InvoiceItemID: movement.InvoiceItemID,
	}
}

func newResponses(movements []model.StockMovement) []response {
	responses := make([]response, 0, len(movements))

	for _, movement := range movements {
		responses = append(responses, newResponse(movement))
	}

	return responses
}

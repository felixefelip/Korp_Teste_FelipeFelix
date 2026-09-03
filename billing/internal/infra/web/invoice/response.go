package invoice

import (
	"time"

	"billing/internal/model"
)

type response struct {
	ID              int            `json:"id"`
	Series          int            `json:"series"`
	Number          int            `json:"number"`
	FormattedNumber string         `json:"formattedNumber"`
	Type            string         `json:"type"`
	Status          string         `json:"status"`
	Items           []itemResponse `json:"items"`
	ProductsTotal   float64        `json:"productsTotal"`
	Total           float64        `json:"total"`
	ICMSBase        float64        `json:"icmsBase"`
	ICMSValue       float64        `json:"icmsValue"`
	IPIBase         float64        `json:"ipiBase"`
	IPIValue        float64        `json:"ipiValue"`

	FailureReason string             `json:"failureReason,omitempty"`
	Shortages     []shortageResponse `json:"shortages,omitempty"`

	ProcessingSince *time.Time `json:"processingSince,omitempty"`
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
		ID:              invoice.ID,
		Series:          invoice.Series,
		Number:          invoice.Number,
		FormattedNumber: invoice.FormattedNumber(),
		Type:            invoice.Type,
		Status:          invoice.Status,
		Items:           items,
		ProductsTotal:   model.RoundMoney(invoice.ProductsTotal()),
		Total:           model.RoundMoney(invoice.Total()),
		ICMSBase:        model.RoundMoney(invoice.ICMSBase()),
		ICMSValue:       model.RoundMoney(invoice.ICMSValue()),
		IPIBase:         model.RoundMoney(invoice.IPIBase()),
		IPIValue:        model.RoundMoney(invoice.IPIValue()),

		FailureReason: invoice.FailureReason,
		Shortages:     newShortageResponses(invoice.Shortages),

		ProcessingSince: invoice.ProcessingSince,
	}
}

func newResponses(invoices []model.Invoice) []response {
	responses := make([]response, 0, len(invoices))

	for _, invoice := range invoices {
		responses = append(responses, newResponse(invoice))
	}

	return responses
}

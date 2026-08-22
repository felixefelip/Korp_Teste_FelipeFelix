package invoice

import (
	"time"

	"billing/internal/model"

	"github.com/google/uuid"
)

const CloseRequestedKey = "invoice.close.requested"

type itemPayload struct {
	InvoiceItemID int `json:"invoiceItemId"`
	ProductID     int `json:"productId"`
	Quantity      int `json:"quantity"`
}

type closeRequestedEvent struct {
	EventID       string        `json:"eventId"`
	OccurredAt    time.Time     `json:"occurredAt"`
	InvoiceID     int           `json:"invoiceId"`
	InvoiceNumber string        `json:"invoiceNumber"`
	Type          string        `json:"type"`
	Items         []itemPayload `json:"items"`
}

func newCloseRequestedEvent(invoice model.Invoice) closeRequestedEvent {
	items := make([]itemPayload, 0, len(invoice.Items))

	for _, item := range invoice.Items {
		items = append(items, itemPayload{
			InvoiceItemID: item.ID,
			ProductID:     item.Product.InventoryID,
			Quantity:      item.Quantity,
		})
	}

	return closeRequestedEvent{
		EventID:       uuid.NewString(),
		OccurredAt:    time.Now().UTC(),
		InvoiceID:     invoice.ID,
		InvoiceNumber: invoice.Number,
		Type:          invoice.Type,
		Items:         items,
	}
}

package invoice

import (
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

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

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleCloseRequested(delivery amqp.Delivery) error {
	var event closeRequestedEvent

	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return err
	}

	fmt.Printf("invoice.close.requested: invoice %d (%s), type %s, event %s\n",
		event.InvoiceID, event.InvoiceNumber, event.Type, event.EventID)

	for _, item := range event.Items {
		fmt.Printf("  item %d: product %d, quantity %d\n",
			item.InvoiceItemID, item.ProductID, item.Quantity)
	}

	return nil
}

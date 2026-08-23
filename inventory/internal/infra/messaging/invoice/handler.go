package invoice

import (
	"encoding/json"
	"fmt"
	"time"

	"inventory/internal/model"
	"inventory/internal/usecase"

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

func (e closeRequestedEvent) toRequest() model.InvoiceStockRequest {
	items := make([]model.InvoiceStockItem, 0, len(e.Items))

	for _, item := range e.Items {
		items = append(items, model.InvoiceStockItem{
			BillingInvoiceItemID: item.InvoiceItemID,
			ProductID:            item.ProductID,
			Quantity:             item.Quantity,
		})
	}

	return model.InvoiceStockRequest{
		InvoiceID:     e.InvoiceID,
		InvoiceNumber: e.InvoiceNumber,
		Type:          e.Type,
		CausationID:   e.EventID,
		Items:         items,
	}
}

type Handler struct {
	usecase usecase.InvoiceStockUsecase
}

func NewHandler(invoiceStockUsecase usecase.InvoiceStockUsecase) *Handler {
	return &Handler{
		usecase: invoiceStockUsecase,
	}
}

func (h *Handler) HandleCloseRequested(delivery amqp.Delivery) error {
	var event closeRequestedEvent

	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return err
	}

	result, err := h.usecase.Apply(event.toRequest())
	if err != nil {
		return err
	}

	fmt.Printf("invoice %d (%s): %s, caused by %s\n",
		event.InvoiceID, event.InvoiceNumber, result.RoutingKey, event.EventID)

	return nil
}

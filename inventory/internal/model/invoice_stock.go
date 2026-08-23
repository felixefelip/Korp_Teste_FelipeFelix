package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	InvoiceCloseRequestedKey = "invoice.close.requested"
	InvoiceStockAppliedKey   = "invoice.stock.applied"
)

type InvoiceStockItem struct {
	InvoiceItemID int
	ProductID     int
	Quantity      int
}

type InvoiceStockRequest struct {
	InvoiceID int
	Type      string
	Items     []InvoiceStockItem
}

func (r InvoiceStockRequest) MovesStockOut() bool {
	return r.Type == InvoiceTypeOut
}

type invoiceStockAppliedPayload struct {
	EventID    string    `json:"eventId"`
	OccurredAt time.Time `json:"occurredAt"`
	InvoiceID  int       `json:"invoiceId"`
}

func NewInvoiceStockApplied(invoiceID int) (OutboxEvent, error) {
	eventID := uuid.NewString()
	occurredAt := time.Now().UTC()

	payload, err := json.Marshal(invoiceStockAppliedPayload{
		EventID:    eventID,
		OccurredAt: occurredAt,
		InvoiceID:  invoiceID,
	})
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		EventID:       eventID,
		AggregateType: OutboxAggregateInvoice,
		AggregateID:   invoiceID,
		RoutingKey:    InvoiceStockAppliedKey,
		Payload:       payload,
		CreatedAt:     occurredAt,
		NextAttemptAt: occurredAt,
	}, nil
}

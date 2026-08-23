package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	InvoiceCloseRequestedKey  = "invoice.close.requested"
	InvoiceReopenRequestedKey = "invoice.reopen.requested"

	InvoiceStockAppliedKey        = "invoice.stock.applied"
	InvoiceStockRejectedKey       = "invoice.stock.rejected"
	InvoiceStockRevertedKey       = "invoice.stock.reverted"
	InvoiceStockRevertRejectedKey = "invoice.stock.revert.rejected"
)

type invoiceItemPayload struct {
	InvoiceItemID int `json:"invoiceItemId"`
	ProductID     int `json:"productId"`
	Quantity      int `json:"quantity"`
}

type reopenRequestedPayload struct {
	EventID       string    `json:"eventId"`
	OccurredAt    time.Time `json:"occurredAt"`
	InvoiceID     int       `json:"invoiceId"`
	InvoiceNumber string    `json:"invoiceNumber"`
}

func NewInvoiceReopenRequested(invoice Invoice) (OutboxEvent, error) {
	eventID := uuid.NewString()
	occurredAt := time.Now().UTC()

	payload, err := json.Marshal(reopenRequestedPayload{
		EventID:       eventID,
		OccurredAt:    occurredAt,
		InvoiceID:     invoice.ID,
		InvoiceNumber: invoice.Number,
	})
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		EventID:       eventID,
		AggregateType: OutboxAggregateInvoice,
		AggregateID:   invoice.ID,
		RoutingKey:    InvoiceReopenRequestedKey,
		Payload:       payload,
		CreatedAt:     occurredAt,
		NextAttemptAt: occurredAt,
	}, nil
}

type invoiceCloseRequestedPayload struct {
	EventID       string               `json:"eventId"`
	OccurredAt    time.Time            `json:"occurredAt"`
	InvoiceID     int                  `json:"invoiceId"`
	InvoiceNumber string               `json:"invoiceNumber"`
	Type          string               `json:"type"`
	Items         []invoiceItemPayload `json:"items"`
}

func NewInvoiceCloseRequested(invoice Invoice) (OutboxEvent, error) {
	items := make([]invoiceItemPayload, 0, len(invoice.Items))

	for _, item := range invoice.Items {
		items = append(items, invoiceItemPayload{
			InvoiceItemID: item.ID,
			ProductID:     item.Product.InventoryID,
			Quantity:      item.Quantity,
		})
	}

	eventID := uuid.NewString()
	occurredAt := time.Now().UTC()

	payload, err := json.Marshal(invoiceCloseRequestedPayload{
		EventID:       eventID,
		OccurredAt:    occurredAt,
		InvoiceID:     invoice.ID,
		InvoiceNumber: invoice.Number,
		Type:          invoice.Type,
		Items:         items,
	})
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		EventID:       eventID,
		AggregateType: OutboxAggregateInvoice,
		AggregateID:   invoice.ID,
		RoutingKey:    InvoiceCloseRequestedKey,
		Payload:       payload,
		CreatedAt:     occurredAt,
		NextAttemptAt: occurredAt,
	}, nil
}

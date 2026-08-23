package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	InvoiceStockAppliedKey  = "invoice.stock.applied"
	InvoiceStockRejectedKey = "invoice.stock.rejected"

	ReasonInsufficientStock = "INSUFFICIENT_STOCK"
	ReasonProductNotFound   = "PRODUCT_NOT_FOUND"
)

type InvoiceStockDecision struct {
	Movements []StockMovement
	Event     OutboxEvent
}

func ResolveInvoiceStock(
	request InvoiceStockRequest,
	products map[int]Product,
	alreadyApplied bool,
) (InvoiceStockDecision, error) {
	if alreadyApplied {
		return appliedDecision(request, nil)
	}

	if request.MissingProducts(products) {
		return rejectedDecision(request, ReasonProductNotFound, nil)
	}

	if shortages := ShortagesFor(request, products); len(shortages) > 0 {
		return rejectedDecision(request, ReasonInsufficientStock, shortages)
	}

	return appliedDecision(request, request.Movements())
}

func appliedDecision(
	request InvoiceStockRequest,
	movements []StockMovement,
) (InvoiceStockDecision, error) {
	event, err := NewInvoiceStockApplied(request)
	if err != nil {
		return InvoiceStockDecision{}, err
	}

	return InvoiceStockDecision{Movements: movements, Event: event}, nil
}

func rejectedDecision(
	request InvoiceStockRequest,
	reason string,
	shortages []StockShortage,
) (InvoiceStockDecision, error) {
	event, err := NewInvoiceStockRejected(request, reason, shortages)
	if err != nil {
		return InvoiceStockDecision{}, err
	}

	return InvoiceStockDecision{Event: event}, nil
}

type invoiceStockResultPayload struct {
	EventID     string          `json:"eventId"`
	CausationID string          `json:"causationId,omitempty"`
	OccurredAt  time.Time       `json:"occurredAt"`
	InvoiceID   int             `json:"invoiceId"`
	Reason      string          `json:"reason,omitempty"`
	Shortages   []StockShortage `json:"shortages,omitempty"`
}

func NewInvoiceStockApplied(request InvoiceStockRequest) (OutboxEvent, error) {
	return newInvoiceStockResult(request, InvoiceStockAppliedKey, "", nil)
}

func NewInvoiceStockRejected(
	request InvoiceStockRequest,
	reason string,
	shortages []StockShortage,
) (OutboxEvent, error) {
	return newInvoiceStockResult(request, InvoiceStockRejectedKey, reason, shortages)
}

func newInvoiceStockResult(
	request InvoiceStockRequest,
	routingKey, reason string,
	shortages []StockShortage,
) (OutboxEvent, error) {
	eventID := uuid.NewString()
	occurredAt := time.Now().UTC()

	payload, err := json.Marshal(invoiceStockResultPayload{
		EventID:     eventID,
		CausationID: request.CausationID,
		OccurredAt:  occurredAt,
		InvoiceID:   request.InvoiceID,
		Reason:      reason,
		Shortages:   shortages,
	})
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		EventID:       eventID,
		CausationID:   request.CausationID,
		AggregateType: OutboxAggregateInvoice,
		AggregateID:   request.InvoiceID,
		RoutingKey:    routingKey,
		Payload:       payload,
		CreatedAt:     occurredAt,
		NextAttemptAt: occurredAt,
	}, nil
}

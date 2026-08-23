package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	InvoiceStockAppliedKey        = "invoice.stock.applied"
	InvoiceStockRejectedKey       = "invoice.stock.rejected"
	InvoiceStockRevertedKey       = "invoice.stock.reverted"
	InvoiceStockRevertRejectedKey = "invoice.stock.revert.rejected"

	ReasonInsufficientStock = "INSUFFICIENT_STOCK"
	ReasonProductNotFound   = "PRODUCT_NOT_FOUND"
	ReasonStockAlreadyUsed  = "STOCK_ALREADY_USED"
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
	return newInvoiceStockResult(
		request.InvoiceID, request.CausationID, InvoiceStockAppliedKey, "", nil,
	)
}

func NewInvoiceStockRejected(
	request InvoiceStockRequest,
	reason string,
	shortages []StockShortage,
) (OutboxEvent, error) {
	return newInvoiceStockResult(
		request.InvoiceID, request.CausationID, InvoiceStockRejectedKey, reason, shortages,
	)
}

func NewInvoiceStockReverted(request InvoiceStockRevertRequest) (OutboxEvent, error) {
	return newInvoiceStockResult(
		request.InvoiceID, request.CausationID, InvoiceStockRevertedKey, "", nil,
	)
}

func NewInvoiceStockRevertRejected(
	request InvoiceStockRevertRequest,
	reason string,
	shortages []StockShortage,
) (OutboxEvent, error) {
	return newInvoiceStockResult(
		request.InvoiceID, request.CausationID, InvoiceStockRevertRejectedKey, reason, shortages,
	)
}

func newInvoiceStockResult(
	invoiceID int,
	causationID string,
	routingKey, reason string,
	shortages []StockShortage,
) (OutboxEvent, error) {
	eventID := uuid.NewString()
	occurredAt := time.Now().UTC()

	payload, err := json.Marshal(invoiceStockResultPayload{
		EventID:     eventID,
		CausationID: causationID,
		OccurredAt:  occurredAt,
		InvoiceID:   invoiceID,
		Reason:      reason,
		Shortages:   shortages,
	})
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		EventID:       eventID,
		CausationID:   causationID,
		AggregateType: OutboxAggregateInvoice,
		AggregateID:   invoiceID,
		RoutingKey:    routingKey,
		Payload:       payload,
		CreatedAt:     occurredAt,
		NextAttemptAt: occurredAt,
	}, nil
}

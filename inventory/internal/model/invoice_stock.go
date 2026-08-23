package model

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	InvoiceCloseRequestedKey = "invoice.close.requested"
	InvoiceStockAppliedKey   = "invoice.stock.applied"
	InvoiceStockRejectedKey  = "invoice.stock.rejected"

	ReasonInsufficientStock = "INSUFFICIENT_STOCK"
	ReasonProductNotFound   = "PRODUCT_NOT_FOUND"
)

type InvoiceStockItem struct {
	BillingInvoiceItemID int
	ProductID            int
	Quantity             int
}

type InvoiceStockRequest struct {
	InvoiceID     int
	InvoiceNumber string
	Type          string
	CausationID   string
	Items         []InvoiceStockItem
}

func (r InvoiceStockRequest) MovesStockOut() bool {
	return r.Type == InvoiceTypeOut
}

func (r InvoiceStockRequest) QuantityRequiredByProduct() map[int]int {
	required := make(map[int]int, len(r.Items))

	for _, item := range r.Items {
		required[item.ProductID] += item.Quantity
	}

	return required
}

func (r InvoiceStockRequest) ProductIDs() []int {
	required := r.QuantityRequiredByProduct()
	ids := make([]int, 0, len(required))

	for productID := range required {
		ids = append(ids, productID)
	}

	sort.Ints(ids)

	return ids
}

func (r InvoiceStockRequest) MissingProducts(products map[int]Product) bool {
	for _, productID := range r.ProductIDs() {
		if _, known := products[productID]; !known {
			return true
		}
	}

	return false
}

func (r InvoiceStockRequest) Movements() []StockMovement {
	movementType := MovementIn
	if r.MovesStockOut() {
		movementType = MovementOut
	}

	movements := make([]StockMovement, 0, len(r.Items))

	for _, item := range r.Items {
		itemID := item.BillingInvoiceItemID
		invoiceID := r.InvoiceID

		movements = append(movements, StockMovement{
			ProductID:            item.ProductID,
			Type:                 movementType,
			Origin:               MovementOriginInvoice,
			Quantity:             item.Quantity,
			Confirmed:            true,
			BillingInvoiceItemID: &itemID,
			BillingInvoiceID:     &invoiceID,
			InvoiceNumber:        r.InvoiceNumber,
			CloseEventID:         r.CausationID,
		})
	}

	return movements
}

type StockShortage struct {
	ProductID int    `json:"productId"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Required  int    `json:"required"`
	Available int    `json:"available"`
}

func ShortagesFor(request InvoiceStockRequest, products map[int]Product) []StockShortage {
	if !request.MovesStockOut() {
		return nil
	}

	quantities := request.QuantityRequiredByProduct()
	shortages := make([]StockShortage, 0)

	for _, productID := range request.ProductIDs() {
		product, known := products[productID]
		if !known {
			continue
		}

		required := quantities[productID]
		if product.Stock >= required {
			continue
		}

		shortages = append(shortages, StockShortage{
			ProductID: productID,
			Code:      product.Code,
			Name:      product.Name,
			Required:  required,
			Available: product.Stock,
		})
	}

	if len(shortages) == 0 {
		return nil
	}

	return shortages
}

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

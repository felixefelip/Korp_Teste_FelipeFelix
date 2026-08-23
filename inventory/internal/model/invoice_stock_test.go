package model_test

import (
	"encoding/json"
	"testing"

	"inventory/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func outRequest(items ...model.InvoiceStockItem) model.InvoiceStockRequest {
	return model.InvoiceStockRequest{
		InvoiceID:     7,
		InvoiceNumber: "NF-0007",
		Type:          model.InvoiceTypeOut,
		CausationID:   "cause-1",
		Items:         items,
	}
}

func TestQuantityRequiredByProductAddsUpItemsOfTheSameProduct(t *testing.T) {
	request := outRequest(
		model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 10},
		model.InvoiceStockItem{BillingInvoiceItemID: 2, ProductID: 42, Quantity: 5},
		model.InvoiceStockItem{BillingInvoiceItemID: 3, ProductID: 43, Quantity: 1},
	)

	assert.Equal(t, map[int]int{42: 15, 43: 1}, request.QuantityRequiredByProduct())
}

func TestProductIDsComeUniqueAndSorted(t *testing.T) {
	request := outRequest(
		model.InvoiceStockItem{ProductID: 43, Quantity: 1},
		model.InvoiceStockItem{ProductID: 42, Quantity: 10},
		model.InvoiceStockItem{ProductID: 42, Quantity: 5},
	)

	assert.Equal(t, []int{42, 43}, request.ProductIDs(),
		"locking always in the same order is what keeps two invoices from deadlocking")
}

func TestShortagesComparesTheSumOfTheItems(t *testing.T) {
	request := outRequest(
		model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 10},
		model.InvoiceStockItem{BillingInvoiceItemID: 2, ProductID: 42, Quantity: 5},
	)
	products := map[int]model.Product{
		42: {ID: 42, Code: "PROD-1", Name: "Parafuso", Stock: 12},
	}

	shortages := model.ShortagesFor(request, products)

	require.Len(t, shortages, 1, "each item fits on its own, the two together do not")
	assert.Equal(t, 15, shortages[0].Required)
	assert.Equal(t, 12, shortages[0].Available)
	assert.Equal(t, "PROD-1", shortages[0].Code)
}

func TestShortagesAreEmptyWhenTheStockCoversTheInvoice(t *testing.T) {
	request := outRequest(model.InvoiceStockItem{ProductID: 42, Quantity: 10})
	products := map[int]model.Product{42: {ID: 42, Stock: 10}}

	assert.Empty(t, model.ShortagesFor(request, products), "an exact balance is enough")
}

func TestAnIncomingInvoiceNeverHasShortages(t *testing.T) {
	request := outRequest(model.InvoiceStockItem{ProductID: 42, Quantity: 10})
	request.Type = model.InvoiceTypeIn
	products := map[int]model.Product{42: {ID: 42, Stock: 0}}

	assert.Empty(t, model.ShortagesFor(request, products), "an inbound invoice only adds")
}

func TestMovementsCarryOneRowPerItem(t *testing.T) {
	request := outRequest(
		model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 10},
		model.InvoiceStockItem{BillingInvoiceItemID: 2, ProductID: 42, Quantity: 5},
	)

	movements := request.Movements()

	require.Len(t, movements, 2, "the item is the grain, so the idempotency key stays unique")
	assert.Equal(t, 1, *movements[0].BillingInvoiceItemID)
	assert.Equal(t, 2, *movements[1].BillingInvoiceItemID)

	for _, movement := range movements {
		assert.Equal(t, model.MovementOut, movement.Type)
		assert.Equal(t, model.MovementOriginInvoice, movement.Origin)
		assert.True(t, movement.Confirmed)
		assert.Equal(t, 7, *movement.BillingInvoiceID)
		assert.Equal(t, "NF-0007", movement.InvoiceNumber)
		assert.Equal(t, "cause-1", movement.CloseEventID)
	}
}

func TestMovementsOfAnIncomingInvoiceAddStock(t *testing.T) {
	request := outRequest(model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 10})
	request.Type = model.InvoiceTypeIn

	assert.Equal(t, model.MovementIn, request.Movements()[0].Type)
}

func TestRejectedEventCarriesTheCauseAndTheShortages(t *testing.T) {
	request := outRequest(model.InvoiceStockItem{ProductID: 42, Quantity: 15})
	shortages := []model.StockShortage{
		{ProductID: 42, Code: "PROD-1", Name: "Parafuso", Required: 15, Available: 12},
	}

	event, err := model.NewInvoiceStockRejected(request, model.ReasonInsufficientStock, shortages)
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRejectedKey, event.RoutingKey)
	assert.Equal(t, "cause-1", event.CausationID)

	var payload struct {
		CausationID string                `json:"causationId"`
		InvoiceID   int                   `json:"invoiceId"`
		Reason      string                `json:"reason"`
		Shortages   []model.StockShortage `json:"shortages"`
	}
	require.NoError(t, json.Unmarshal(event.Payload, &payload))

	assert.Equal(t, "cause-1", payload.CausationID)
	assert.Equal(t, 7, payload.InvoiceID)
	assert.Equal(t, model.ReasonInsufficientStock, payload.Reason)
	assert.Equal(t, shortages, payload.Shortages)
}

func TestAppliedEventCarriesTheCauseAndNoReason(t *testing.T) {
	event, err := model.NewInvoiceStockApplied(outRequest())
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockAppliedKey, event.RoutingKey)
	assert.Equal(t, "cause-1", event.CausationID)
	assert.NotContains(t, string(event.Payload), "reason")
}

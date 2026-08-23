package model_test

import (
	"encoding/json"
	"testing"

	"inventory/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func available(stock int) map[int]model.Product {
	return map[int]model.Product{
		42: {ID: 42, Code: "PROD-1", Name: "Parafuso", Stock: stock},
	}
}

func reasonOf(t *testing.T, event model.OutboxEvent) string {
	t.Helper()

	var payload struct {
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(event.Payload, &payload))

	return payload.Reason
}

func TestResolveWritesOneMovementPerItemWhenTheStockCovers(t *testing.T) {
	request := outRequest(
		model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 10},
		model.InvoiceStockItem{BillingInvoiceItemID: 2, ProductID: 42, Quantity: 5},
	)

	decision, err := model.ResolveInvoiceStock(request, available(20), false)
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockAppliedKey, decision.Event.RoutingKey)
	assert.Len(t, decision.Movements, 2)
}

func TestResolveRefusesAndWritesNothingWhenTheSumExceedsTheStock(t *testing.T) {
	request := outRequest(
		model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 10},
		model.InvoiceStockItem{BillingInvoiceItemID: 2, ProductID: 42, Quantity: 5},
	)

	decision, err := model.ResolveInvoiceStock(request, available(12), false)
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRejectedKey, decision.Event.RoutingKey)
	assert.Equal(t, model.ReasonInsufficientStock, reasonOf(t, decision.Event))
	assert.Empty(t, decision.Movements, "all or nothing")
}

func TestResolveRefusesWhenAProductIsGone(t *testing.T) {
	request := outRequest(model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 99, Quantity: 1})

	decision, err := model.ResolveInvoiceStock(request, available(20), false)
	require.NoError(t, err)

	assert.Equal(t, model.ReasonProductNotFound, reasonOf(t, decision.Event))
	assert.Empty(t, decision.Movements)
}

func TestResolveTellsProductGoneApartFromMissingStock(t *testing.T) {
	request := outRequest(
		model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 999},
		model.InvoiceStockItem{BillingInvoiceItemID: 2, ProductID: 99, Quantity: 1},
	)

	decision, err := model.ResolveInvoiceStock(request, available(12), false)
	require.NoError(t, err)

	assert.Equal(t, model.ReasonProductNotFound, reasonOf(t, decision.Event),
		"a product that no longer exists is not a balance problem")
}

func TestResolveAcceptsAnIncomingInvoiceWithNoStockAtAll(t *testing.T) {
	request := outRequest(model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 7})
	request.Type = model.InvoiceTypeIn

	decision, err := model.ResolveInvoiceStock(request, available(0), false)
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockAppliedKey, decision.Event.RoutingKey)
	assert.Len(t, decision.Movements, 1)
}

func TestResolveAnswersAppliedAgainWithoutWritingWhenItAlreadyRan(t *testing.T) {
	request := outRequest(model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 10})

	decision, err := model.ResolveInvoiceStock(request, available(20), true)
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockAppliedKey, decision.Event.RoutingKey,
		"a redelivery must not leave the invoice hanging")
	assert.Empty(t, decision.Movements, "the stock was already taken")
}

func TestResolvePutsTheReplayCheckBeforeTheBalanceCheck(t *testing.T) {
	request := outRequest(model.InvoiceStockItem{BillingInvoiceItemID: 1, ProductID: 42, Quantity: 999})

	decision, err := model.ResolveInvoiceStock(request, available(0), true)
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockAppliedKey, decision.Event.RoutingKey,
		"the stock it once took is already gone from the balance it is being compared against")
}

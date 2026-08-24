package model_test

import (
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func catalog() []model.Product {
	return []model.Product{
		{InventoryID: 1, Code: "PRD-0001", Name: "Notebook Dell", Unit: "UN", Price: 4500, Active: true},
		{InventoryID: 2, Code: "PRD-0002", Name: "Notebook HP", Unit: "UN", Price: 3900, Active: true},
		{InventoryID: 3, Code: "PRD-0003", Name: "Monitor LG 24", Unit: "UN", Price: 890.5, Active: true},
		{InventoryID: 4, Code: "PRD-0004", Name: "Café em grãos", Unit: "KG", Price: 52.9, Active: true},
		{InventoryID: 5, Code: "PRD-0005", Name: "Teclado antigo", Unit: "UN", Price: 99, Active: false},
	}
}

func extraction(items ...model.ExtractedItem) model.InvoiceDraftExtraction {
	return model.InvoiceDraftExtraction{Type: model.InvoiceTypeOut, Items: items}
}

func TestResolveInvoiceDraftMatchesByCodeFromTheModel(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "aquele notebook", Code: "PRD-0002", Quantity: 2}),
		catalog(),
	)

	require.Len(t, draft.Items, 1)
	assert.Empty(t, draft.Unresolved)
	assert.Equal(t, "PRD-0002", draft.Items[0].ProductCode)
	assert.Equal(t, 2, draft.Items[0].Quantity)
}

func TestResolveInvoiceDraftFallsBackWhenTheCodeDoesNotExist(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "notebook dell", Code: "PRD-9999", Quantity: 1}),
		catalog(),
	)

	require.Len(t, draft.Items, 1)
	assert.Equal(t, "PRD-0001", draft.Items[0].ProductCode)
}

func TestResolveInvoiceDraftReportsNotFoundWhenOnlyTheHallucinatedCodeIsGiven(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "cadeira gamer", Code: "PRD-9999", Quantity: 1}),
		catalog(),
	)

	assert.Empty(t, draft.Items)
	require.Len(t, draft.Unresolved, 1)
	assert.Equal(t, model.DraftNotFound, draft.Unresolved[0].Reason)
	assert.Equal(t, "cadeira gamer", draft.Unresolved[0].Text)
	assert.Empty(t, draft.Unresolved[0].Candidates)
}

func TestResolveInvoiceDraftMatchesCodeWrittenInTheText(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "prd-0003", Quantity: 3}),
		catalog(),
	)

	require.Len(t, draft.Items, 1)
	assert.Equal(t, "PRD-0003", draft.Items[0].ProductCode)
}

func TestResolveInvoiceDraftMatchesPluralAndCase(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "NOTEBOOKS DELL", Quantity: 5}),
		catalog(),
	)

	require.Len(t, draft.Items, 1)
	assert.Equal(t, "PRD-0001", draft.Items[0].ProductCode)
}

func TestResolveInvoiceDraftMatchesIgnoringAccents(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "cafe em graos", Quantity: 2}),
		catalog(),
	)

	require.Len(t, draft.Items, 1)
	assert.Equal(t, "PRD-0004", draft.Items[0].ProductCode)
}

func TestResolveInvoiceDraftReportsAmbiguityWithCandidates(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "notebook", Quantity: 1}),
		catalog(),
	)

	assert.Empty(t, draft.Items)
	require.Len(t, draft.Unresolved, 1)
	assert.Equal(t, model.DraftAmbiguous, draft.Unresolved[0].Reason)
	require.Len(t, draft.Unresolved[0].Candidates, 2)
	assert.Equal(t, "PRD-0001", draft.Unresolved[0].Candidates[0].Code)
	assert.Equal(t, "PRD-0002", draft.Unresolved[0].Candidates[1].Code)
}

func TestResolveInvoiceDraftIgnoresInactiveProducts(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "teclado antigo", Code: "PRD-0005", Quantity: 1}),
		catalog(),
	)

	assert.Empty(t, draft.Items)
	require.Len(t, draft.Unresolved, 1)
	assert.Equal(t, model.DraftNotFound, draft.Unresolved[0].Reason)
}

func TestResolveInvoiceDraftRejectsQuantityWithoutValue(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "notebook dell", Code: "PRD-0001", Quantity: 0}),
		catalog(),
	)

	assert.Empty(t, draft.Items)
	require.Len(t, draft.Unresolved, 1)
	assert.Equal(t, model.DraftInvalidQuantity, draft.Unresolved[0].Reason)
}

func TestResolveInvoiceDraftTakesPriceAndUnitFromTheCatalog(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "café em grãos a 10 reais o quilo", Code: "PRD-0004", Quantity: 3}),
		catalog(),
	)

	require.Len(t, draft.Items, 1)
	assert.Equal(t, 52.9, draft.Items[0].UnitPrice)
	assert.Equal(t, "KG", draft.Items[0].Unit)
	assert.Equal(t, "Café em grãos", draft.Items[0].ProductName)
	assert.Equal(t, 4, draft.Items[0].Product.InventoryID)
}

func TestResolveInvoiceDraftKeepsRepeatedProductsAsSeparateItems(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(
			model.ExtractedItem{Text: "notebook dell", Code: "PRD-0001", Quantity: 2},
			model.ExtractedItem{Text: "notebook dell", Code: "PRD-0001", Quantity: 3},
		),
		catalog(),
	)

	require.Len(t, draft.Items, 2)
	assert.Equal(t, 2, draft.Items[0].Quantity)
	assert.Equal(t, 3, draft.Items[1].Quantity)
}

func TestResolveInvoiceDraftKeepsResolvedItemsWhenOneFails(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(
			model.ExtractedItem{Text: "notebook dell", Code: "PRD-0001", Quantity: 2},
			model.ExtractedItem{Text: "cadeira gamer", Quantity: 1},
		),
		catalog(),
	)

	require.Len(t, draft.Items, 1)
	require.Len(t, draft.Unresolved, 1)
	assert.Equal(t, "cadeira gamer", draft.Unresolved[0].Text)
}

func TestResolveInvoiceDraftReadsTheType(t *testing.T) {
	tests := map[string]string{
		"IN":      model.InvoiceTypeIn,
		"in":      model.InvoiceTypeIn,
		"OUT":     model.InvoiceTypeOut,
		"":        model.InvoiceTypeOut,
		"INVALID": model.InvoiceTypeOut,
	}

	for value, expected := range tests {
		draft := model.ResolveInvoiceDraft(model.InvoiceDraftExtraction{Type: value}, catalog())

		assert.Equal(t, expected, draft.Type, "type %q", value)
	}
}

func TestResolveInvoiceDraftWithoutCatalog(t *testing.T) {
	draft := model.ResolveInvoiceDraft(
		extraction(model.ExtractedItem{Text: "notebook dell", Quantity: 1}),
		nil,
	)

	assert.Empty(t, draft.Items)
	require.Len(t, draft.Unresolved, 1)
	assert.Equal(t, model.DraftNotFound, draft.Unresolved[0].Reason)
}

package model_test

import (
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestResolveInvoiceSeriesPrefersWhatWasAsked(t *testing.T) {
	assert.Equal(t, 7, model.ResolveInvoiceSeries(7, 2))
}

func TestResolveInvoiceSeriesFallsBackToTheLastUsed(t *testing.T) {
	assert.Equal(t, 2, model.ResolveInvoiceSeries(0, 2))
}

func TestResolveInvoiceSeriesStartsAtOneWhenThereIsNoInvoice(t *testing.T) {
	assert.Equal(t, model.FirstInvoiceSeries, model.ResolveInvoiceSeries(0, 0))
}

func TestNextInvoiceDocumentStartsAtOne(t *testing.T) {
	document := model.NextInvoiceDocument(1, 0)

	assert.Equal(t, 1, document.Series)
	assert.Equal(t, 1, document.Number)
	assert.True(t, document.Suggested())
}

func TestNextInvoiceDocumentFollowsTheLastNumber(t *testing.T) {
	document := model.NextInvoiceDocument(3, 57)

	assert.Equal(t, 3, document.Series)
	assert.Equal(t, 58, document.Number)
}

func TestNextInvoiceDocumentSuggestsNothingWhenTheSeriesIsFull(t *testing.T) {
	document := model.NextInvoiceDocument(1, model.MaxInvoiceNumber)

	assert.Equal(t, 1, document.Series)
	assert.False(t, document.Suggested())
	assert.Zero(t, document.Number)
}

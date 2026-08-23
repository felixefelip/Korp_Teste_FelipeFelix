package usecase_test

import (
	"encoding/json"
	"testing"

	"inventory/internal/model"
	"inventory/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func invoiceStockRequest() model.InvoiceStockRequest {
	return model.InvoiceStockRequest{
		InvoiceID: 7,
		Type:      model.InvoiceTypeOut,
		Items: []model.InvoiceStockItem{
			{InvoiceItemID: 3, ProductID: 42, Quantity: 10},
		},
	}
}

func TestApplyRecordsTheStockAppliedEvent(t *testing.T) {
	repository := &fakeMovementRepository{}
	invoiceStockUsecase := usecase.NewInvoiceStockUsecase(repository)

	require.NoError(t, invoiceStockUsecase.Apply(invoiceStockRequest()))

	event := repository.recordedEvent
	assert.Equal(t, model.InvoiceStockAppliedKey, event.RoutingKey)
	assert.Equal(t, model.OutboxAggregateInvoice, event.AggregateType)
	assert.Equal(t, 7, event.AggregateID)
	assert.NotEmpty(t, event.EventID)
	assert.False(t, event.Published(), "it is the relay that publishes")
}

func TestApplyHandsTheWholeRequestToTheRepository(t *testing.T) {
	repository := &fakeMovementRepository{}
	invoiceStockUsecase := usecase.NewInvoiceStockUsecase(repository)

	require.NoError(t, invoiceStockUsecase.Apply(invoiceStockRequest()))

	assert.Equal(t, invoiceStockRequest(), repository.receivedRequest,
		"the movements and the event are written together")
}

func TestApplyPayloadCarriesTheInvoice(t *testing.T) {
	repository := &fakeMovementRepository{}
	invoiceStockUsecase := usecase.NewInvoiceStockUsecase(repository)

	require.NoError(t, invoiceStockUsecase.Apply(invoiceStockRequest()))

	var payload struct {
		EventID   string `json:"eventId"`
		InvoiceID int    `json:"invoiceId"`
	}
	require.NoError(t, json.Unmarshal(repository.recordedEvent.Payload, &payload))

	assert.Equal(t, 7, payload.InvoiceID)
	assert.Equal(t, repository.recordedEvent.EventID, payload.EventID)
}

func TestApplyPropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeMovementRepository{err: errRepository}
	invoiceStockUsecase := usecase.NewInvoiceStockUsecase(repository)

	assert.ErrorIs(t, invoiceStockUsecase.Apply(invoiceStockRequest()), errRepository)
}

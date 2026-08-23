package usecase_test

import (
	"testing"

	"inventory/internal/model"
	"inventory/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func invoiceStockRequest() model.InvoiceStockRequest {
	return model.InvoiceStockRequest{
		InvoiceID:     7,
		InvoiceNumber: "NF-0007",
		Type:          model.InvoiceTypeOut,
		CausationID:   "cause-1",
		Items: []model.InvoiceStockItem{
			{BillingInvoiceItemID: 3, ProductID: 42, Quantity: 10},
		},
	}
}

func TestApplyHandsTheWholeRequestToTheRepository(t *testing.T) {
	repository := &fakeMovementRepository{}
	invoiceStockUsecase := usecase.NewInvoiceStockUsecase(repository)

	_, err := invoiceStockUsecase.Apply(invoiceStockRequest())

	require.NoError(t, err)
	assert.Equal(t, invoiceStockRequest(), repository.receivedRequest,
		"validating and writing happen together, under the same lock")
}

func TestApplyReturnsTheResultTheLedgerRecorded(t *testing.T) {
	repository := &fakeMovementRepository{
		recordedEvent: model.OutboxEvent{RoutingKey: model.InvoiceStockRejectedKey},
	}
	invoiceStockUsecase := usecase.NewInvoiceStockUsecase(repository)

	result, err := invoiceStockUsecase.Apply(invoiceStockRequest())

	require.NoError(t, err)
	assert.Equal(t, model.InvoiceStockRejectedKey, result.RoutingKey)
}

func TestApplyPropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeMovementRepository{err: errRepository}
	invoiceStockUsecase := usecase.NewInvoiceStockUsecase(repository)

	_, err := invoiceStockUsecase.Apply(invoiceStockRequest())

	assert.ErrorIs(t, err, errRepository)
}

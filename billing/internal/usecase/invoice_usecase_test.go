package usecase_test

import (
	"encoding/json"
	"errors"
	"testing"

	"billing/internal/model"
	"billing/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	newID    int
	invoice  model.Invoice
	invoices []model.Invoice
	err      error

	receivedID        int
	receivedInvoice   model.Invoice
	closedID          int
	recordedEvent     model.OutboxEvent
	confirmedID       int
	rejectedID        int
	rejectedReason    string
	rejectedShortages []model.InvoiceShortage
	transitionMoved   bool
	reopenedID        int
	retriedID         int
	retriedStatus     string
	deletedID         int
	calls             int
}

func (f *fakeRepository) GetInvoices() ([]model.Invoice, error) {
	f.calls++
	return f.invoices, f.err
}

func (f *fakeRepository) GetInvoiceByID(id int) (model.Invoice, error) {
	f.calls++
	f.receivedID = id
	return f.invoice, f.err
}

func (f *fakeRepository) CreateInvoice(invoice model.Invoice) (int, error) {
	f.calls++
	f.receivedInvoice = invoice
	return f.newID, f.err
}

func (f *fakeRepository) UpdateInvoice(invoice model.Invoice) error {
	f.calls++
	f.receivedInvoice = invoice
	return f.err
}

func (f *fakeRepository) CloseInvoice(id int, event model.OutboxEvent) error {
	f.calls++
	f.closedID = id
	f.recordedEvent = event
	return f.err
}

func (f *fakeRepository) ConfirmClose(id int) (bool, error) {
	f.calls++
	f.confirmedID = id
	return f.transitionMoved, f.err
}

func (f *fakeRepository) RejectClose(
	id int,
	reason string,
	shortages []model.InvoiceShortage,
) (bool, error) {
	f.calls++
	f.rejectedID = id
	f.rejectedReason = reason
	f.rejectedShortages = shortages
	return f.transitionMoved, f.err
}

func (f *fakeRepository) ConfirmReopen(id int) (bool, error) {
	f.calls++
	f.confirmedID = id
	return f.transitionMoved, f.err
}

func (f *fakeRepository) RejectReopen(
	id int,
	reason string,
	shortages []model.InvoiceShortage,
) (bool, error) {
	f.calls++
	f.rejectedID = id
	f.rejectedReason = reason
	f.rejectedShortages = shortages
	return f.transitionMoved, f.err
}

func (f *fakeRepository) ReopenInvoice(id int, event model.OutboxEvent) error {
	f.calls++
	f.reopenedID = id
	f.recordedEvent = event
	return f.err
}

func (f *fakeRepository) RetryInvoice(id int, status string, event model.OutboxEvent) error {
	f.calls++
	f.retriedID = id
	f.retriedStatus = status
	f.recordedEvent = event
	return f.err
}

func (f *fakeRepository) DeleteInvoice(id int) error {
	f.calls++
	f.deletedID = id
	return f.err
}

var errRepository = errors.New("database down")

func newUsecase(repository model.InvoiceRepository) usecase.InvoiceUsecase {
	return usecase.NewInvoiceUsecase(repository)
}

func TestGetInvoicesReturnsWhatTheRepositoryHolds(t *testing.T) {
	stored := []model.Invoice{
		{ID: 1, Series: 1, Number: 1, Status: "OPEN"},
		{ID: 2, Series: 1, Number: 2, Status: "CLOSED"},
	}
	repository := &fakeRepository{invoices: stored}
	invoiceUsecase := newUsecase(repository)

	invoices, err := invoiceUsecase.GetInvoices()

	require.NoError(t, err)
	assert.Equal(t, stored, invoices)
	assert.Equal(t, 1, repository.calls)
}

func TestGetInvoicesPropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeRepository{err: errRepository}
	invoiceUsecase := newUsecase(repository)

	invoices, err := invoiceUsecase.GetInvoices()

	require.ErrorIs(t, err, errRepository)
	assert.Empty(t, invoices)
}

func TestGetInvoiceByIDForwardsTheID(t *testing.T) {
	repository := &fakeRepository{invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: "OPEN"}}
	invoiceUsecase := newUsecase(repository)

	invoice, err := invoiceUsecase.GetInvoiceByID(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.receivedID)
	assert.Equal(t, 7, invoice.Number)
}

func TestGetInvoiceByIDPropagatesTheRepositoryError(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{err: errRepository})

	_, err := invoiceUsecase.GetInvoiceByID(7)

	assert.ErrorIs(t, err, errRepository)
}

func TestCreateInvoiceReturnsWhatWasStored(t *testing.T) {
	stored := model.Invoice{ID: 7, Series: 1, Number: 1, Status: "OPEN"}
	repository := &fakeRepository{newID: 7, invoice: stored}
	invoiceUsecase := newUsecase(repository)

	created, err := invoiceUsecase.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})

	require.NoError(t, err)
	assert.Equal(t, stored, created, "the response must be read back from the repository")
	assert.Equal(t, 7, repository.receivedID, "it reads back the id the repository generated")
}

func TestCreateInvoiceHandsTheRepositoryWhatItReceived(t *testing.T) {
	repository := &fakeRepository{newID: 1}
	invoiceUsecase := newUsecase(repository)

	invoice := model.Invoice{
		Series: 1, Number: 2,
		Status: "CLOSED",
		Items: []model.InvoiceItem{
			{ProductCode: "PRD-0001", ProductName: "Camiseta", Unit: "UN", Quantity: 2, UnitPrice: 30.99},
		},
	}
	_, err := invoiceUsecase.CreateInvoice(invoice)

	require.NoError(t, err)
	assert.Equal(t, invoice, repository.receivedInvoice, "the items ride along untouched")
}

func TestCreateInvoicePropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeRepository{err: errRepository}
	invoiceUsecase := newUsecase(repository)

	created, err := invoiceUsecase.CreateInvoice(model.Invoice{Series: 1, Number: 1, Status: "OPEN"})

	require.ErrorIs(t, err, errRepository)
	assert.Zero(t, created, "on failure nothing partially filled should leak out")
}

func TestUpdateInvoiceReturnsWhatWasStored(t *testing.T) {
	stored := model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusOpen}
	repository := &fakeRepository{invoice: stored}
	invoiceUsecase := newUsecase(repository)

	invoice := model.Invoice{ID: 7, Series: 1, Number: 7}
	updated, err := invoiceUsecase.UpdateInvoice(invoice)

	require.NoError(t, err)
	assert.Equal(t, stored, updated, "the response must be read back from the repository")
	assert.Equal(t, invoice, repository.receivedInvoice)
	assert.Equal(t, 7, repository.receivedID)
}

func TestUpdateInvoiceHandsTheItemsToTheRepository(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusOpen},
	}
	invoiceUsecase := newUsecase(repository)

	items := []model.InvoiceItem{
		{ProductCode: "PRD-0001", ProductName: "Camiseta", Unit: "UN", Quantity: 2, UnitPrice: 30.99},
	}

	_, err := invoiceUsecase.UpdateInvoice(model.Invoice{ID: 7, Series: 1, Number: 7, Status: "OPEN", Items: items})

	require.NoError(t, err)
	assert.Equal(t, items, repository.receivedInvoice.Items)
}

func TestUpdateInvoicePropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeRepository{err: errRepository}
	invoiceUsecase := newUsecase(repository)

	updated, err := invoiceUsecase.UpdateInvoice(model.Invoice{ID: 7, Series: 1, Number: 7, Status: "OPEN"})

	require.ErrorIs(t, err, errRepository)
	assert.Zero(t, updated, "on failure nothing partially filled should leak out")
}

func TestDeleteInvoiceRemovesAnOpenOne(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusOpen},
	}
	invoiceUsecase := newUsecase(repository)

	err := invoiceUsecase.DeleteInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.deletedID)
}

func TestDeleteInvoiceRefusesAClosedOne(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusClosed},
	}
	invoiceUsecase := newUsecase(repository)

	err := invoiceUsecase.DeleteInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceClosed)
	assert.Zero(t, repository.deletedID, "it stops at the read, without deleting")
	assert.Equal(t, 1, repository.calls)
}

func TestDeleteInvoiceWhenMissingPropagatesTheError(t *testing.T) {
	repository := &fakeRepository{err: errRepository}
	invoiceUsecase := newUsecase(repository)

	err := invoiceUsecase.DeleteInvoice(7)

	require.ErrorIs(t, err, errRepository)
	assert.Zero(t, repository.deletedID)
}

func TestCloseInvoiceClosesAnOpenOne(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusOpen},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.closedID)
}

func TestCloseInvoiceRefusesOneAlreadyClosed(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusClosed},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceClosed)
	assert.Zero(t, repository.closedID, "it stops at the read, without writing")
	assert.Equal(t, 1, repository.calls)
}

func TestCloseInvoiceReturnsTheInvoiceAsItEndedUp(t *testing.T) {
	stored := model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusOpen}
	repository := &fakeRepository{invoice: stored}
	invoiceUsecase := newUsecase(repository)

	closed, err := invoiceUsecase.CloseInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, stored, closed, "it re-reads instead of guessing the new state")
}

func TestCloseInvoiceWhenMissingPropagatesTheError(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{err: errRepository})

	_, err := invoiceUsecase.CloseInvoice(7)

	require.ErrorIs(t, err, errRepository)
}

func TestCloseInvoiceRecordsTheCloseRequestedEvent(t *testing.T) {
	stored := model.Invoice{
		ID:     7,
		Series: 1, Number: 7,
		Type:   model.InvoiceTypeOut,
		Status: model.InvoiceStatusOpen,
		Items: []model.InvoiceItem{
			{ID: 3, Quantity: 10, Product: model.Product{InventoryID: 42}},
		},
	}
	repository := &fakeRepository{invoice: stored}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.NoError(t, err)

	event := repository.recordedEvent
	assert.Equal(t, model.InvoiceCloseRequestedKey, event.RoutingKey)
	assert.Equal(t, model.OutboxAggregateInvoice, event.AggregateType)
	assert.Equal(t, 7, event.AggregateID)
	assert.NotEmpty(t, event.EventID)
	assert.False(t, event.Published(), "the relay is the one that publishes it")
}

func TestCloseInvoicePayloadCarriesTheInventoryProductID(t *testing.T) {
	stored := model.Invoice{
		ID:     7,
		Series: 1, Number: 7,
		Type:   model.InvoiceTypeOut,
		Status: model.InvoiceStatusOpen,
		Items: []model.InvoiceItem{
			{ID: 3, ProductID: 1, Quantity: 10, Product: model.Product{InventoryID: 42}},
		},
	}
	repository := &fakeRepository{invoice: stored}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.CloseInvoice(7)
	require.NoError(t, err)

	var payload struct {
		InvoiceID int `json:"invoiceId"`
		Items     []struct {
			InvoiceItemID int `json:"invoiceItemId"`
			ProductID     int `json:"productId"`
			Quantity      int `json:"quantity"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(repository.recordedEvent.Payload, &payload))

	assert.Equal(t, 7, payload.InvoiceID)
	require.Len(t, payload.Items, 1)
	assert.Equal(t, 3, payload.Items[0].InvoiceItemID)
	assert.Equal(t, 42, payload.Items[0].ProductID, "the inventory identity, not the local replica one")
	assert.Equal(t, 10, payload.Items[0].Quantity)
}

func TestCloseInvoiceRecordsNothingWhenAlreadyClosed(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusClosed},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceClosed)
	assert.Empty(t, repository.recordedEvent.RoutingKey)
}

func TestReopenInvoiceReopensAClosedOne(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusClosed},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.ReopenInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.reopenedID)
}

func TestReopenInvoiceRefusesOneAlreadyOpen(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusOpen},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.ReopenInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceOpen)
	assert.Zero(t, repository.reopenedID, "it stops at the read, without writing")
	assert.Equal(t, 1, repository.calls)
}

func TestReopenInvoiceWhenMissingPropagatesTheError(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{err: errRepository})

	_, err := invoiceUsecase.ReopenInvoice(7)

	require.ErrorIs(t, err, errRepository)
}

func TestUpdateInvoiceRefusesAClosedOne(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusClosed},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.UpdateInvoice(model.Invoice{ID: 7, Series: 1, Number: 9})

	require.ErrorIs(t, err, model.ErrInvoiceClosed)
	assert.Zero(t, repository.receivedInvoice, "it stops at the read, without writing")
	assert.Equal(t, 1, repository.calls)
}

func processingInvoice() model.Invoice {
	return model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusClosing}
}

func TestUpdateInvoiceRefusesOneBeingProcessed(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{invoice: processingInvoice()})

	_, err := invoiceUsecase.UpdateInvoice(model.Invoice{ID: 7, Series: 1, Number: 7})

	require.ErrorIs(t, err, model.ErrInvoiceProcessing)
}

func TestDeleteInvoiceRefusesOneBeingProcessed(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{invoice: processingInvoice()})

	require.ErrorIs(t, invoiceUsecase.DeleteInvoice(7), model.ErrInvoiceProcessing)
}

func TestCloseInvoiceRefusesOneAlreadyBeingProcessed(t *testing.T) {
	repository := &fakeRepository{invoice: processingInvoice()}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceProcessing)
	assert.Empty(t, repository.recordedEvent.RoutingKey, "it does not ask for the stock twice")
}

func TestReopenInvoiceRefusesOneBeingProcessed(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{invoice: processingInvoice()})

	_, err := invoiceUsecase.ReopenInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceProcessing)
}

func TestConfirmCloseMovesTheInvoice(t *testing.T) {
	repository := &fakeRepository{transitionMoved: true}
	invoiceUsecase := newUsecase(repository)

	require.NoError(t, invoiceUsecase.ConfirmClose(7))
	assert.Equal(t, 7, repository.confirmedID)
}

func TestConfirmCloseReportsWhenTheInvoiceMovedAlready(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{transitionMoved: false})

	err := invoiceUsecase.ConfirmClose(7)

	require.ErrorIs(t, err, model.ErrInvoiceNotProcessing,
		"a repeated result must be recognised, not applied twice")
}

func TestRejectCloseCarriesTheReason(t *testing.T) {
	repository := &fakeRepository{transitionMoved: true}
	invoiceUsecase := newUsecase(repository)

	require.NoError(t, invoiceUsecase.RejectClose(7, "INSUFFICIENT_STOCK", nil))

	assert.Equal(t, 7, repository.rejectedID)
	assert.Equal(t, "INSUFFICIENT_STOCK", repository.rejectedReason)
}

func TestRejectCloseReportsWhenTheInvoiceMovedAlready(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{transitionMoved: false})

	require.ErrorIs(t, invoiceUsecase.RejectClose(7, "INSUFFICIENT_STOCK", nil), model.ErrInvoiceNotProcessing)
}

func TestRejectCloseCarriesTheShortagesToTheRepository(t *testing.T) {
	repository := &fakeRepository{transitionMoved: true}
	invoiceUsecase := newUsecase(repository)

	shortages := []model.InvoiceShortage{
		{InventoryID: 42, ProductCode: "PROD-1", ProductName: "Parafuso", Required: 50, Available: 42},
	}

	require.NoError(t, invoiceUsecase.RejectClose(7, "INSUFFICIENT_STOCK", shortages))

	assert.Equal(t, shortages, repository.rejectedShortages)
}

func TestRetryInvoiceRepublishesTheCloseRequest(t *testing.T) {
	stored := model.Invoice{
		ID:     7,
		Series: 1, Number: 7,
		Type:   model.InvoiceTypeOut,
		Status: model.InvoiceStatusClosing,
		Items: []model.InvoiceItem{
			{ID: 3, ProductID: 1, Quantity: 10, Product: model.Product{InventoryID: 42}},
		},
	}
	repository := &fakeRepository{invoice: stored}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.RetryInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.retriedID)
	assert.Equal(t, model.InvoiceStatusClosing, repository.retriedStatus, "the status is the guard, not a transition")
	assert.Equal(t, model.InvoiceCloseRequestedKey, repository.recordedEvent.RoutingKey)
}

func TestRetryInvoiceRepublishesTheReopenRequest(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusReopening},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.RetryInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, model.InvoiceStatusReopening, repository.retriedStatus)
	assert.Equal(t, model.InvoiceReopenRequestedKey, repository.recordedEvent.RoutingKey)
}

func TestRetryInvoiceRefusesOneThatIsNotBeingProcessed(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Series: 1, Number: 7, Status: model.InvoiceStatusOpen},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.RetryInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceNotProcessing)
	assert.Zero(t, repository.retriedID, "it stops at the read, without writing")
	assert.Equal(t, 1, repository.calls)
}

func TestRetryInvoiceWhenMissingPropagatesTheError(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{err: errRepository})

	_, err := invoiceUsecase.RetryInvoice(7)

	assert.ErrorIs(t, err, errRepository)
}

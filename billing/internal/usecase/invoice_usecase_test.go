package usecase_test

import (
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

	receivedID      int
	receivedInvoice model.Invoice
	closedID        int
	reopenedID      int
	deletedID       int
	calls           int
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

func (f *fakeRepository) CloseInvoice(id int) error {
	f.calls++
	f.closedID = id
	return f.err
}

func (f *fakeRepository) ReopenInvoice(id int) error {
	f.calls++
	f.reopenedID = id
	return f.err
}

func (f *fakeRepository) DeleteInvoice(id int) error {
	f.calls++
	f.deletedID = id
	return f.err
}

type fakePublisher struct {
	published model.Invoice
	err       error
	calls     int
}

func (f *fakePublisher) PublishCloseRequested(invoice model.Invoice) error {
	f.calls++
	f.published = invoice

	return f.err
}

var (
	errRepository = errors.New("database down")
	errPublisher  = errors.New("broker down")
)

func newUsecase(repository model.InvoiceRepository) usecase.InvoiceUsecase {
	return usecase.NewInvoiceUsecase(repository, &fakePublisher{})
}

func newUsecaseWithPublisher(
	repository model.InvoiceRepository,
	publisher model.InvoiceEventPublisher,
) usecase.InvoiceUsecase {
	return usecase.NewInvoiceUsecase(repository, publisher)
}

func TestGetInvoicesReturnsWhatTheRepositoryHolds(t *testing.T) {
	stored := []model.Invoice{
		{ID: 1, Number: "NF-0001", Status: "OPEN"},
		{ID: 2, Number: "NF-0002", Status: "CLOSED"},
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
	repository := &fakeRepository{invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: "OPEN"}}
	invoiceUsecase := newUsecase(repository)

	invoice, err := invoiceUsecase.GetInvoiceByID(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.receivedID)
	assert.Equal(t, "NF-0007", invoice.Number)
}

func TestGetInvoiceByIDPropagatesTheRepositoryError(t *testing.T) {
	invoiceUsecase := newUsecase(&fakeRepository{err: errRepository})

	_, err := invoiceUsecase.GetInvoiceByID(7)

	assert.ErrorIs(t, err, errRepository)
}

func TestCreateInvoiceReturnsWhatWasStored(t *testing.T) {
	stored := model.Invoice{ID: 7, Number: "NF-0001", Status: "OPEN"}
	repository := &fakeRepository{newID: 7, invoice: stored}
	invoiceUsecase := newUsecase(repository)

	created, err := invoiceUsecase.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})

	require.NoError(t, err)
	assert.Equal(t, stored, created, "the response must be read back from the repository")
	assert.Equal(t, 7, repository.receivedID, "it reads back the id the repository generated")
}

func TestCreateInvoiceHandsTheRepositoryWhatItReceived(t *testing.T) {
	repository := &fakeRepository{newID: 1}
	invoiceUsecase := newUsecase(repository)

	invoice := model.Invoice{
		Number: "NF-0002",
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

	created, err := invoiceUsecase.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})

	require.ErrorIs(t, err, errRepository)
	assert.Zero(t, created, "on failure nothing partially filled should leak out")
}

func TestUpdateInvoiceReturnsWhatWasStored(t *testing.T) {
	stored := model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusOpen}
	repository := &fakeRepository{invoice: stored}
	invoiceUsecase := newUsecase(repository)

	invoice := model.Invoice{ID: 7, Number: "NF-0007"}
	updated, err := invoiceUsecase.UpdateInvoice(invoice)

	require.NoError(t, err)
	assert.Equal(t, stored, updated, "the response must be read back from the repository")
	assert.Equal(t, invoice, repository.receivedInvoice)
	assert.Equal(t, 7, repository.receivedID)
}

func TestUpdateInvoiceHandsTheItemsToTheRepository(t *testing.T) {
	repository := &fakeRepository{}
	invoiceUsecase := newUsecase(repository)

	items := []model.InvoiceItem{
		{ProductCode: "PRD-0001", ProductName: "Camiseta", Unit: "UN", Quantity: 2, UnitPrice: 30.99},
	}

	_, err := invoiceUsecase.UpdateInvoice(model.Invoice{ID: 7, Number: "NF-0007", Status: "OPEN", Items: items})

	require.NoError(t, err)
	assert.Equal(t, items, repository.receivedInvoice.Items)
}

func TestUpdateInvoicePropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeRepository{err: errRepository}
	invoiceUsecase := newUsecase(repository)

	updated, err := invoiceUsecase.UpdateInvoice(model.Invoice{ID: 7, Number: "NF-0007", Status: "OPEN"})

	require.ErrorIs(t, err, errRepository)
	assert.Zero(t, updated, "on failure nothing partially filled should leak out")
}

func TestDeleteInvoiceRemovesAnOpenOne(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusOpen},
	}
	invoiceUsecase := newUsecase(repository)

	err := invoiceUsecase.DeleteInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.deletedID)
}

func TestDeleteInvoiceRefusesAClosedOne(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusClosed},
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
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusOpen},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.closedID)
}

func TestCloseInvoiceRefusesOneAlreadyClosed(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusClosed},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceClosed)
	assert.Zero(t, repository.closedID, "it stops at the read, without writing")
	assert.Equal(t, 1, repository.calls)
}

func TestCloseInvoiceReturnsTheInvoiceAsItEndedUp(t *testing.T) {
	stored := model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusOpen}
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

func TestCloseInvoicePublishesTheClosedInvoice(t *testing.T) {
	stored := model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusOpen}
	repository := &fakeRepository{invoice: stored}
	publisher := &fakePublisher{}
	invoiceUsecase := newUsecaseWithPublisher(repository, publisher)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, 1, publisher.calls)
	assert.Equal(t, stored, publisher.published, "it publishes what the repository holds after closing")
}

func TestCloseInvoiceRefusesWhenPublishingFails(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusOpen},
	}
	invoiceUsecase := newUsecaseWithPublisher(repository, &fakePublisher{err: errPublisher})

	_, err := invoiceUsecase.CloseInvoice(7)

	require.ErrorIs(t, err, errPublisher)
}

func TestCloseInvoiceSkipsPublishingWhenAlreadyClosed(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusClosed},
	}
	publisher := &fakePublisher{}
	invoiceUsecase := newUsecaseWithPublisher(repository, publisher)

	_, err := invoiceUsecase.CloseInvoice(7)

	require.ErrorIs(t, err, model.ErrInvoiceClosed)
	assert.Zero(t, publisher.calls)
}

func TestReopenInvoiceReopensAClosedOne(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusClosed},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.ReopenInvoice(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.reopenedID)
}

func TestReopenInvoiceRefusesOneAlreadyOpen(t *testing.T) {
	repository := &fakeRepository{
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusOpen},
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
		invoice: model.Invoice{ID: 7, Number: "NF-0007", Status: model.InvoiceStatusClosed},
	}
	invoiceUsecase := newUsecase(repository)

	_, err := invoiceUsecase.UpdateInvoice(model.Invoice{ID: 7, Number: "NF-0009"})

	require.ErrorIs(t, err, model.ErrInvoiceClosed)
	assert.Zero(t, repository.receivedInvoice, "it stops at the read, without writing")
	assert.Equal(t, 1, repository.calls)
}

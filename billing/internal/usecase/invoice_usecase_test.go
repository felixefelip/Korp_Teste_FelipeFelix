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

var errRepository = errors.New("database down")

func newUsecase(repository model.InvoiceRepository) usecase.InvoiceUsecase {
	return usecase.NewInvoiceUsecase(repository)
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
	stored := model.Invoice{ID: 7, Number: "NF-0007", Status: "CLOSED"}
	repository := &fakeRepository{invoice: stored}
	invoiceUsecase := newUsecase(repository)

	invoice := model.Invoice{ID: 7, Number: "NF-0007", Status: "CLOSED"}
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

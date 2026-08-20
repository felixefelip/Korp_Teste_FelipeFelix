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
	invoices []model.Invoice
	err      error

	receivedInvoice model.Invoice
	calls           int
}

func (f *fakeRepository) GetInvoices() ([]model.Invoice, error) {
	f.calls++
	return f.invoices, f.err
}

func (f *fakeRepository) CreateInvoice(invoice model.Invoice) (int, error) {
	f.calls++
	f.receivedInvoice = invoice
	return f.newID, f.err
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

func TestCreateInvoiceReturnsTheInvoiceWithTheGeneratedID(t *testing.T) {
	repository := &fakeRepository{newID: 7}
	invoiceUsecase := newUsecase(repository)

	created, err := invoiceUsecase.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})

	require.NoError(t, err)
	assert.Equal(t, 7, created.ID)
	assert.Equal(t, "NF-0001", created.Number)
	assert.Equal(t, "OPEN", created.Status)
	assert.Equal(t, 1, repository.calls)
}

func TestCreateInvoiceHandsTheRepositoryWhatItReceived(t *testing.T) {
	repository := &fakeRepository{newID: 1}
	invoiceUsecase := newUsecase(repository)

	invoice := model.Invoice{Number: "NF-0002", Status: "CLOSED"}
	_, err := invoiceUsecase.CreateInvoice(invoice)

	require.NoError(t, err)
	assert.Equal(t, invoice, repository.receivedInvoice)
}

func TestCreateInvoicePropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeRepository{err: errRepository}
	invoiceUsecase := newUsecase(repository)

	created, err := invoiceUsecase.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})

	require.ErrorIs(t, err, errRepository)
	assert.Zero(t, created, "on failure nothing partially filled should leak out")
}

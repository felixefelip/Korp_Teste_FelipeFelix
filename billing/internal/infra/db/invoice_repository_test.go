package db_test

import (
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGetInvoicesReturnsEverythingStored(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "CLOSED"})
	require.NoError(t, err)

	invoices, err := repository.GetInvoices()
	require.NoError(t, err)

	require.Len(t, invoices, 2)
	assert.ElementsMatch(t,
		[]string{"NF-0001", "NF-0002"},
		[]string{invoices[0].Number, invoices[1].Number},
	)
}

func TestGetInvoicesReturnsTheStoredFields(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "CLOSED"})
	require.NoError(t, err)

	invoices, err := repository.GetInvoices()
	require.NoError(t, err)

	require.Len(t, invoices, 1)
	assert.Equal(t, model.Invoice{ID: id, Number: "NF-0001", Status: "CLOSED"}, invoices[0])
}

func TestGetInvoicesWithNothingStoredReturnsAnEmptyList(t *testing.T) {
	repository := newRepository(t)

	invoices, err := repository.GetInvoices()

	require.NoError(t, err)
	assert.Empty(t, invoices)
}

func TestGetInvoiceByIDReturnsTheInvoice(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	invoice, err := repository.GetInvoiceByID(id)

	require.NoError(t, err)
	assert.Equal(t, model.Invoice{ID: id, Number: "NF-0001", Status: "OPEN"}, invoice)
}

func TestGetInvoiceByIDWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.GetInvoiceByID(404)

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCreateInvoice(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)
	assert.NotZero(t, id, "the database should have generated an id")

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "NF-0001", saved.Number)
	assert.Equal(t, "OPEN", saved.Status)
}

func TestCreateInvoiceKeepsTheClosedStatus(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "CLOSED"})
	require.NoError(t, err)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "CLOSED", saved.Status)
}

func TestCreateInvoiceGeneratesSequentialIDs(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	second, err := repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "OPEN"})
	require.NoError(t, err)

	assert.Greater(t, second, first, "each invoice should get its own id")
}

func TestCreateInvoiceAcceptsADuplicateNumber(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	_, err = repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err, "a duplicate number must be accepted")
}

func TestUpdateInvoiceChangesEveryField(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	err = repository.UpdateInvoice(model.Invoice{ID: id, Number: "NF-0002", Status: "CLOSED"})
	require.NoError(t, err)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, model.Invoice{ID: id, Number: "NF-0002", Status: "CLOSED"}, saved)
}

func TestUpdateInvoiceWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	err := repository.UpdateInvoice(model.Invoice{ID: 9999, Number: "NF-0001", Status: "OPEN"})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUpdateInvoiceLeavesTheOtherInvoicesAlone(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateInvoice(model.Invoice{Number: "NF-0001", Status: "OPEN"})
	require.NoError(t, err)

	second, err := repository.CreateInvoice(model.Invoice{Number: "NF-0002", Status: "OPEN"})
	require.NoError(t, err)

	require.NoError(t, repository.UpdateInvoice(model.Invoice{ID: first, Number: "NF-0001", Status: "CLOSED"}))

	var untouched model.Invoice
	require.NoError(t, testConnection.First(&untouched, second).Error)

	assert.Equal(t, "NF-0002", untouched.Number)
	assert.Equal(t, "OPEN", untouched.Status)
}

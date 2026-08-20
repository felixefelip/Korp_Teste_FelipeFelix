package db_test

import (
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

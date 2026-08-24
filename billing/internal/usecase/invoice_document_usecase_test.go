package usecase_test

import (
	"errors"
	"testing"

	"billing/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextDocumentCountsFromTheAskedSeries(t *testing.T) {
	repository := &fakeRepository{lastSeries: 1, lastNumber: 12}
	invoiceUsecase := usecase.NewInvoiceUsecase(repository)

	document, err := invoiceUsecase.NextDocument(3)

	require.NoError(t, err)
	assert.Equal(t, 3, document.Series)
	assert.Equal(t, 13, document.Number)
	assert.Equal(t, 3, repository.receivedSeries)
}

func TestNextDocumentUsesTheLastSeriesWhenNoneIsAsked(t *testing.T) {
	repository := &fakeRepository{lastSeries: 2, lastNumber: 5}
	invoiceUsecase := usecase.NewInvoiceUsecase(repository)

	document, err := invoiceUsecase.NextDocument(0)

	require.NoError(t, err)
	assert.Equal(t, 2, document.Series)
	assert.Equal(t, 6, document.Number)
	assert.Equal(t, 2, repository.receivedSeries)
}

func TestNextDocumentOnAnEmptyDatabase(t *testing.T) {
	repository := &fakeRepository{}
	invoiceUsecase := usecase.NewInvoiceUsecase(repository)

	document, err := invoiceUsecase.NextDocument(0)

	require.NoError(t, err)
	assert.Equal(t, 1, document.Series)
	assert.Equal(t, 1, document.Number)
}

func TestNextDocumentWhenTheDatabaseFails(t *testing.T) {
	failure := errors.New("database down")
	repository := &fakeRepository{err: failure}
	invoiceUsecase := usecase.NewInvoiceUsecase(repository)

	_, err := invoiceUsecase.NextDocument(0)

	assert.ErrorIs(t, err, failure)
}

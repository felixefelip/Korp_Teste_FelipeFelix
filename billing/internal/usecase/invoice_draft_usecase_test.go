package usecase_test

import (
	"context"
	"errors"
	"testing"

	"billing/internal/model"
	"billing/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCatalogRepository struct {
	products []model.Product
	err      error
}

func (f *fakeCatalogRepository) GetProducts() ([]model.Product, error) {
	return f.products, f.err
}

func (f *fakeCatalogRepository) UpsertProduct(model.Product) error {
	return nil
}

func (f *fakeCatalogRepository) DeactivateProduct(int) error {
	return nil
}

type fakeExtractor struct {
	extraction model.InvoiceDraftExtraction
	err        error

	receivedPrompt  string
	receivedCatalog []model.Product
	calls           int
}

func (f *fakeExtractor) Extract(
	_ context.Context,
	prompt string,
	catalog []model.Product,
) (model.InvoiceDraftExtraction, error) {
	f.calls++
	f.receivedPrompt = prompt
	f.receivedCatalog = catalog

	return f.extraction, f.err
}

func draftCatalog() []model.Product {
	return []model.Product{
		{InventoryID: 1, Code: "PRD-0001", Name: "Notebook Dell", Unit: "UN", Price: 4500, Active: true},
	}
}

func TestDraftInvoiceResolvesTheExtractionAgainstTheCatalog(t *testing.T) {
	repository := &fakeCatalogRepository{products: draftCatalog()}
	extractor := &fakeExtractor{
		extraction: model.InvoiceDraftExtraction{
			Type:  model.InvoiceTypeOut,
			Items: []model.ExtractedItem{{Text: "notebook dell", Quantity: 2}},
		},
	}

	draftUsecase := usecase.NewInvoiceDraftUsecase(repository, extractor)

	draft, err := draftUsecase.DraftInvoice(context.Background(), "vender 2 notebooks dell")

	require.NoError(t, err)
	assert.Equal(t, model.InvoiceTypeOut, draft.Type)
	require.Len(t, draft.Items, 1)
	assert.Equal(t, "PRD-0001", draft.Items[0].ProductCode)
	assert.Equal(t, 4500.0, draft.Items[0].UnitPrice)
	assert.Equal(t, "vender 2 notebooks dell", extractor.receivedPrompt)
	assert.Equal(t, draftCatalog(), extractor.receivedCatalog)
}

func TestDraftInvoiceWithoutExtractorConfigured(t *testing.T) {
	repository := &fakeCatalogRepository{products: draftCatalog()}

	draftUsecase := usecase.NewInvoiceDraftUsecase(repository, nil)

	_, err := draftUsecase.DraftInvoice(context.Background(), "vender 2 notebooks dell")

	assert.ErrorIs(t, err, model.ErrDraftUnavailable)
}

func TestDraftInvoiceWhenTheCatalogFails(t *testing.T) {
	failure := errors.New("database down")
	repository := &fakeCatalogRepository{err: failure}
	extractor := &fakeExtractor{}

	draftUsecase := usecase.NewInvoiceDraftUsecase(repository, extractor)

	_, err := draftUsecase.DraftInvoice(context.Background(), "vender 2 notebooks dell")

	assert.ErrorIs(t, err, failure)
	assert.Zero(t, extractor.calls)
}

func TestDraftInvoiceWhenTheExtractionFails(t *testing.T) {
	failure := errors.New("upstream unavailable")
	repository := &fakeCatalogRepository{products: draftCatalog()}
	extractor := &fakeExtractor{err: failure}

	draftUsecase := usecase.NewInvoiceDraftUsecase(repository, extractor)

	_, err := draftUsecase.DraftInvoice(context.Background(), "vender 2 notebooks dell")

	assert.ErrorIs(t, err, failure)
}

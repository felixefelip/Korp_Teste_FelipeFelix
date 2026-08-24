package usecase

import (
	"context"

	"billing/internal/model"
)

type InvoiceDraftUsecase struct {
	repository model.ProductRepository
	extractor  model.InvoiceDraftExtractor
}

func NewInvoiceDraftUsecase(
	repository model.ProductRepository,
	extractor model.InvoiceDraftExtractor,
) InvoiceDraftUsecase {
	return InvoiceDraftUsecase{
		repository: repository,
		extractor:  extractor,
	}
}

func (du *InvoiceDraftUsecase) DraftInvoice(
	ctx context.Context,
	prompt string,
) (model.InvoiceDraft, error) {
	if du.extractor == nil {
		return model.InvoiceDraft{}, model.ErrDraftUnavailable
	}

	catalog, err := du.repository.GetProducts()
	if err != nil {
		return model.InvoiceDraft{}, err
	}

	extraction, err := du.extractor.Extract(ctx, prompt, catalog)
	if err != nil {
		return model.InvoiceDraft{}, err
	}

	return model.ResolveInvoiceDraft(extraction, catalog), nil
}

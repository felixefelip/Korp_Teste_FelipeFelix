package model

import "context"

type InvoiceDraftExtractor interface {
	Extract(ctx context.Context, prompt string, catalog []Product) (InvoiceDraftExtraction, error)
}

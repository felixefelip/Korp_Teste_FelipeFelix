package invoice

import (
	"billing/internal/model"
)

type documentResponse struct {
	Series int  `json:"series"`
	Number *int `json:"number"`
}

func newDocumentResponse(document model.InvoiceDocument) documentResponse {
	response := documentResponse{Series: document.Series}

	if document.Suggested() {
		number := document.Number
		response.Number = &number
	}

	return response
}

package invoice

import (
	"billing/internal/model"
)

type draftItemResponse struct {
	InventoryID int     `json:"inventoryId"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Unit        string  `json:"unit"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	ICMSRate    float64 `json:"icmsRate"`
	IPIRate     float64 `json:"ipiRate"`
}

type candidateResponse struct {
	InventoryID int    `json:"inventoryId"`
	Code        string `json:"code"`
	Name        string `json:"name"`
}

type unresolvedResponse struct {
	Text       string              `json:"text"`
	Quantity   int                 `json:"quantity"`
	Reason     string              `json:"reason"`
	Candidates []candidateResponse `json:"candidates"`
}

type draftResponse struct {
	Type       string               `json:"type"`
	Items      []draftItemResponse  `json:"items"`
	Unresolved []unresolvedResponse `json:"unresolved"`
}

func newDraftResponse(draft model.InvoiceDraft) draftResponse {
	response := draftResponse{
		Type:       draft.Type,
		Items:      make([]draftItemResponse, 0, len(draft.Items)),
		Unresolved: make([]unresolvedResponse, 0, len(draft.Unresolved)),
	}

	for _, item := range draft.Items {
		response.Items = append(response.Items, draftItemResponse{
			InventoryID: item.Product.InventoryID,
			Code:        item.ProductCode,
			Name:        item.ProductName,
			Unit:        item.Unit,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			ICMSRate:    item.ICMSRate,
			IPIRate:     item.IPIRate,
		})
	}

	for _, unresolved := range draft.Unresolved {
		response.Unresolved = append(response.Unresolved, unresolvedResponse{
			Text:       unresolved.Text,
			Quantity:   unresolved.Quantity,
			Reason:     unresolved.Reason,
			Candidates: newCandidateResponses(unresolved.Candidates),
		})
	}

	return response
}

func newCandidateResponses(products []model.Product) []candidateResponse {
	candidates := make([]candidateResponse, 0, len(products))

	for _, product := range products {
		candidates = append(candidates, candidateResponse{
			InventoryID: product.InventoryID,
			Code:        product.Code,
			Name:        product.Name,
		})
	}

	return candidates
}

package invoice

import (
	"strings"

	"billing/internal/model"
)

type createRequest struct {
	Number string        `json:"number" binding:"required,max=30"`
	Type   string        `json:"type"   binding:"required,oneof=IN OUT"`
	Items  []itemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r createRequest) toModel() model.Invoice {
	return model.Invoice{
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Type:   r.Type,
		Status: model.InvoiceStatusOpen,
		Items:  toItemModels(r.Items),
	}
}

type updateRequest struct {
	Number string        `json:"number" binding:"required,max=30"`
	Items  []itemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r updateRequest) toModel(id int) model.Invoice {
	return model.Invoice{
		ID:     id,
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Items:  toItemModels(r.Items),
	}
}

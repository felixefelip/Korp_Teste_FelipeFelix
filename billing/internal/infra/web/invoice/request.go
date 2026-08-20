package invoice

import (
	"strings"

	"billing/internal/model"
)

type createRequest struct {
	Number string        `json:"number" binding:"required,max=30"`
	Status string        `json:"status" binding:"required,oneof=OPEN CLOSED"`
	Items  []itemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r createRequest) toModel() model.Invoice {
	return model.Invoice{
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Status: r.Status,
		Items:  toItemModels(r.Items),
	}
}

type updateRequest struct {
	Number string        `json:"number" binding:"required,max=30"`
	Status string        `json:"status" binding:"required,oneof=OPEN CLOSED"`
	Items  []itemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r updateRequest) toModel(id int) model.Invoice {
	return model.Invoice{
		ID:     id,
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Status: r.Status,
		Items:  toItemModels(r.Items),
	}
}

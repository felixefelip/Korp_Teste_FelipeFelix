package invoice

import (
	"billing/internal/model"
)

type createRequest struct {
	Series *int          `json:"series" binding:"required,gt=0,lte=999"`
	Number *int          `json:"number" binding:"required,gt=0,lte=999999"`
	Type   string        `json:"type"   binding:"required,oneof=IN OUT"`
	Items  []itemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r createRequest) toModel() model.Invoice {
	return model.Invoice{
		Series: *r.Series,
		Number: *r.Number,
		Type:   r.Type,
		Status: model.InvoiceStatusOpen,
		Items:  toItemModels(r.Items),
	}
}

type updateRequest struct {
	Series *int          `json:"series" binding:"required,gt=0,lte=999"`
	Number *int          `json:"number" binding:"required,gt=0,lte=999999"`
	Items  []itemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r updateRequest) toModel(id int) model.Invoice {
	return model.Invoice{
		ID:     id,
		Series: *r.Series,
		Number: *r.Number,
		Items:  toItemModels(r.Items),
	}
}

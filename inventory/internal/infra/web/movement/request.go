package movement

import (
	"inventory/internal/model"
)

type createRequest struct {
	Type      string `json:"type"      binding:"required,oneof=in out"`
	Quantity  *int   `json:"quantity"  binding:"required,gt=0"`
	Confirmed bool   `json:"confirmed"`
}

func (r createRequest) toModel(productID int) model.StockMovement {
	return model.StockMovement{
		ProductID: productID,
		Type:      r.Type,
		Quantity:  *r.Quantity,
		Confirmed: r.Confirmed,
	}
}

type updateRequest struct {
	Type      string `json:"type"      binding:"required,oneof=in out"`
	Quantity  *int   `json:"quantity"  binding:"required,gt=0"`
	Confirmed bool   `json:"confirmed"`
}

func (r updateRequest) toModel(id int) model.StockMovement {
	return model.StockMovement{
		ID:        id,
		Type:      r.Type,
		Quantity:  *r.Quantity,
		Confirmed: r.Confirmed,
	}
}

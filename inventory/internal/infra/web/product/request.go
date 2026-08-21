package product

import (
	"strings"

	"inventory/internal/model"
)

type createRequest struct {
	Code  string   `json:"code"  binding:"required,max=30"`
	Name  string   `json:"name"  binding:"required,max=255"`
	Unit  string   `json:"unit"  binding:"required,oneof=UN CX PC KG L M"`
	Price *float64 `json:"price" binding:"required,gte=0"`
	Stock int      `json:"stock" binding:"gte=0"`
}

func (r createRequest) toModel() model.Product {
	return model.Product{
		Code:  strings.ToUpper(strings.TrimSpace(r.Code)),
		Name:  strings.TrimSpace(r.Name),
		Unit:  r.Unit,
		Price: *r.Price,
		Stock: r.Stock,
	}
}

type updateRequest struct {
	Code  string   `json:"code"  binding:"required,max=30"`
	Name  string   `json:"name"  binding:"required,max=255"`
	Unit  string   `json:"unit"  binding:"required,oneof=UN CX PC KG L M"`
	Price *float64 `json:"price" binding:"required,gte=0"`
}

func (r updateRequest) toModel(id int) model.Product {
	return model.Product{
		ID:    id,
		Code:  strings.ToUpper(strings.TrimSpace(r.Code)),
		Name:  strings.TrimSpace(r.Name),
		Unit:  r.Unit,
		Price: *r.Price,
	}
}

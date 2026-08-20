package web

import (
	"strings"

	"inventory/internal/model"
)

type createProductRequest struct {
	Code  string   `json:"code"  binding:"required,max=30"`
	Name  string   `json:"name"  binding:"required,max=255"`
	Unit  string   `json:"unit"  binding:"required,oneof=UN CX PC KG L M"`
	Price *float64 `json:"price" binding:"required,gte=0"`
	Stock int      `json:"stock" binding:"gte=0"`
}

func (r createProductRequest) toModel() model.Product {
	return model.Product{
		Code:  strings.ToUpper(strings.TrimSpace(r.Code)),
		Name:  strings.TrimSpace(r.Name),
		Unit:  r.Unit,
		Price: *r.Price,
		Stock: r.Stock,
	}
}

type productResponse struct {
	ID    int     `json:"id"`
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Unit  string  `json:"unit"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func newProductResponse(product model.Product) productResponse {
	return productResponse{
		ID:    product.ID,
		Code:  product.Code,
		Name:  product.Name,
		Unit:  product.Unit,
		Price: product.Price,
		Stock: product.Stock,
	}
}

func newProductResponses(products []model.Product) []productResponse {
	responses := make([]productResponse, 0, len(products))

	for _, product := range products {
		responses = append(responses, newProductResponse(product))
	}

	return responses
}

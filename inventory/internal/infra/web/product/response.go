package product

import (
	"inventory/internal/model"
)

type response struct {
	ID    int     `json:"id"`
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Unit  string  `json:"unit"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func newResponse(product model.Product) response {
	return response{
		ID:    product.ID,
		Code:  product.Code,
		Name:  product.Name,
		Unit:  product.Unit,
		Price: product.Price,
		Stock: product.Stock,
	}
}

func newResponses(products []model.Product) []response {
	responses := make([]response, 0, len(products))

	for _, product := range products {
		responses = append(responses, newResponse(product))
	}

	return responses
}

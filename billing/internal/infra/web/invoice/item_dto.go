package invoice

import (
	"strings"

	"billing/internal/model"
)

type itemRequest struct {
	InventoryID int      `json:"inventoryId" binding:"required,gt=0"`
	Code        string   `json:"code"        binding:"required,max=30"`
	Name        string   `json:"name"        binding:"required,max=255"`
	Unit        string   `json:"unit"        binding:"required,max=10"`
	Quantity    *int     `json:"quantity"    binding:"required,gt=0"`
	UnitPrice   *float64 `json:"unitPrice"   binding:"required,gte=0"`
}

func (r itemRequest) toModel() model.InvoiceItem {
	code := strings.ToUpper(strings.TrimSpace(r.Code))
	name := strings.TrimSpace(r.Name)
	unit := strings.ToUpper(strings.TrimSpace(r.Unit))

	return model.InvoiceItem{
		ProductCode: code,
		ProductName: name,
		Unit:        unit,
		Quantity:    *r.Quantity,
		UnitPrice:   *r.UnitPrice,
		Product: model.Product{
			InventoryID: r.InventoryID,
			Code:        code,
			Name:        name,
			Unit:        unit,
			Price:       *r.UnitPrice,
		},
	}
}

func toItemModels(items []itemRequest) []model.InvoiceItem {
	if items == nil {
		return nil
	}

	models := make([]model.InvoiceItem, 0, len(items))

	for _, item := range items {
		models = append(models, item.toModel())
	}

	return models
}

type itemResponse struct {
	ID          int     `json:"id"`
	ProductID   int     `json:"productId"`
	InventoryID int     `json:"inventoryId"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Unit        string  `json:"unit"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Total       float64 `json:"total"`
}

func newItemResponse(item model.InvoiceItem) itemResponse {
	return itemResponse{
		ID:          item.ID,
		ProductID:   item.ProductID,
		InventoryID: item.Product.InventoryID,
		Code:        item.ProductCode,
		Name:        item.ProductName,
		Unit:        item.Unit,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		Total:       round(item.Total()),
	}
}

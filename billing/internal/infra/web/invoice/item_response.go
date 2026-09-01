package invoice

import (
	"billing/internal/model"
)

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
	ICMSRate    float64 `json:"icmsRate"`
	ICMSBase    float64 `json:"icmsBase"`
	ICMSValue   float64 `json:"icmsValue"`
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
		Total:       model.RoundMoney(item.Total()),
		ICMSRate:    item.ICMSRate,
		ICMSBase:    item.ICMSBase,
		ICMSValue:   item.ICMSValue,
	}
}

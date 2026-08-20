package web

import (
	"math"
	"strings"

	"billing/internal/model"
)

type invoiceItemRequest struct {
	InventoryID int      `json:"inventoryId" binding:"required,gt=0"`
	Code        string   `json:"code"        binding:"required,max=30"`
	Name        string   `json:"name"        binding:"required,max=255"`
	Unit        string   `json:"unit"        binding:"required,max=10"`
	Quantity    *int     `json:"quantity"    binding:"required,gt=0"`
	UnitPrice   *float64 `json:"unitPrice"   binding:"required,gte=0"`
}

func (r invoiceItemRequest) toModel() model.InvoiceItem {
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

func toItemModels(items []invoiceItemRequest) []model.InvoiceItem {
	if items == nil {
		return nil
	}

	models := make([]model.InvoiceItem, 0, len(items))

	for _, item := range items {
		models = append(models, item.toModel())
	}

	return models
}

type createInvoiceRequest struct {
	Number string               `json:"number" binding:"required,max=30"`
	Status string               `json:"status" binding:"required,oneof=OPEN CLOSED"`
	Items  []invoiceItemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r createInvoiceRequest) toModel() model.Invoice {
	return model.Invoice{
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Status: r.Status,
		Items:  toItemModels(r.Items),
	}
}

type updateInvoiceRequest struct {
	Number string               `json:"number" binding:"required,max=30"`
	Status string               `json:"status" binding:"required,oneof=OPEN CLOSED"`
	Items  []invoiceItemRequest `json:"items"  binding:"omitempty,dive"`
}

func (r updateInvoiceRequest) toModel(id int) model.Invoice {
	return model.Invoice{
		ID:     id,
		Number: strings.ToUpper(strings.TrimSpace(r.Number)),
		Status: r.Status,
		Items:  toItemModels(r.Items),
	}
}

type invoiceItemResponse struct {
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

func newInvoiceItemResponse(item model.InvoiceItem) invoiceItemResponse {
	return invoiceItemResponse{
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

type invoiceResponse struct {
	ID     int                   `json:"id"`
	Number string                `json:"number"`
	Status string                `json:"status"`
	Items  []invoiceItemResponse `json:"items"`
	Total  float64               `json:"total"`
}

func newInvoiceResponse(invoice model.Invoice) invoiceResponse {
	items := make([]invoiceItemResponse, 0, len(invoice.Items))

	for _, item := range invoice.Items {
		items = append(items, newInvoiceItemResponse(item))
	}

	return invoiceResponse{
		ID:     invoice.ID,
		Number: invoice.Number,
		Status: invoice.Status,
		Items:  items,
		Total:  round(invoice.Total()),
	}
}

func newInvoiceResponses(invoices []model.Invoice) []invoiceResponse {
	responses := make([]invoiceResponse, 0, len(invoices))

	for _, invoice := range invoices {
		responses = append(responses, newInvoiceResponse(invoice))
	}

	return responses
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

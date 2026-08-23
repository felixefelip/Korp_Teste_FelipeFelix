package model

import "sort"

const (
	InvoiceCloseRequestedKey  = "invoice.close.requested"
	InvoiceReopenRequestedKey = "invoice.reopen.requested"
)

type InvoiceStockItem struct {
	BillingInvoiceItemID int
	ProductID            int
	Quantity             int
}

type InvoiceStockRequest struct {
	InvoiceID     int
	InvoiceNumber string
	Type          string
	CausationID   string
	Items         []InvoiceStockItem
}

func (r InvoiceStockRequest) MovesStockOut() bool {
	return r.Type == InvoiceTypeOut
}

func (r InvoiceStockRequest) QuantityRequiredByProduct() map[int]int {
	required := make(map[int]int, len(r.Items))

	for _, item := range r.Items {
		required[item.ProductID] += item.Quantity
	}

	return required
}

func (r InvoiceStockRequest) ProductIDs() []int {
	return sortedProductIDs(r.QuantityRequiredByProduct())
}

func sortedProductIDs(byProduct map[int]int) []int {
	ids := make([]int, 0, len(byProduct))

	for productID := range byProduct {
		ids = append(ids, productID)
	}

	sort.Ints(ids)

	return ids
}

func (r InvoiceStockRequest) MissingProducts(products map[int]Product) bool {
	for _, productID := range r.ProductIDs() {
		if _, known := products[productID]; !known {
			return true
		}
	}

	return false
}

func (r InvoiceStockRequest) Movements() []StockMovement {
	movementType := MovementIn
	if r.MovesStockOut() {
		movementType = MovementOut
	}

	movements := make([]StockMovement, 0, len(r.Items))

	for _, item := range r.Items {
		itemID := item.BillingInvoiceItemID
		invoiceID := r.InvoiceID

		movements = append(movements, StockMovement{
			ProductID:            item.ProductID,
			Type:                 movementType,
			Origin:               MovementOriginInvoice,
			Quantity:             item.Quantity,
			Confirmed:            true,
			BillingInvoiceItemID: &itemID,
			BillingInvoiceID:     &invoiceID,
			InvoiceNumber:        r.InvoiceNumber,
			CloseEventID:         r.CausationID,
		})
	}

	return movements
}

type StockShortage struct {
	ProductID int    `json:"productId"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Required  int    `json:"required"`
	Available int    `json:"available"`
}

func ShortagesFor(request InvoiceStockRequest, products map[int]Product) []StockShortage {
	if !request.MovesStockOut() {
		return nil
	}

	quantities := request.QuantityRequiredByProduct()
	shortages := make([]StockShortage, 0)

	for _, productID := range request.ProductIDs() {
		product, known := products[productID]
		if !known {
			continue
		}

		required := quantities[productID]
		if product.Stock >= required {
			continue
		}

		shortages = append(shortages, StockShortage{
			ProductID: productID,
			Code:      product.Code,
			Name:      product.Name,
			Required:  required,
			Available: product.Stock,
		})
	}

	if len(shortages) == 0 {
		return nil
	}

	return shortages
}

package model

type InvoiceStockRevertRequest struct {
	InvoiceID     int
	InvoiceNumber string
	CausationID   string
}

type InvoiceRevertDecision struct {
	Movements []StockMovement
	Event     OutboxEvent
}

func BalanceAfterRemoving(movements []StockMovement, products map[int]Product) map[int]int {
	balances := make(map[int]int, len(products))

	for productID, product := range products {
		balances[productID] = product.Stock
	}

	for _, movement := range movements {
		if movement.Type == MovementIn {
			balances[movement.ProductID] -= movement.Quantity

			continue
		}

		balances[movement.ProductID] += movement.Quantity
	}

	return balances
}

func RevertShortagesFor(movements []StockMovement, products map[int]Product) []StockShortage {
	balances := BalanceAfterRemoving(movements, products)
	shortages := make([]StockShortage, 0)

	for _, productID := range sortedProductIDs(balances) {
		if balances[productID] >= 0 {
			continue
		}

		product := products[productID]

		shortages = append(shortages, StockShortage{
			ProductID: productID,
			Code:      product.Code,
			Name:      product.Name,
			Required:  product.Stock - balances[productID],
			Available: product.Stock,
		})
	}

	if len(shortages) == 0 {
		return nil
	}

	return shortages
}

func ResolveInvoiceRevert(
	request InvoiceStockRevertRequest,
	movements []StockMovement,
	products map[int]Product,
) (InvoiceRevertDecision, error) {
	if len(movements) == 0 {
		event, err := NewInvoiceStockReverted(request)

		return InvoiceRevertDecision{Event: event}, err
	}

	if shortages := RevertShortagesFor(movements, products); len(shortages) > 0 {
		event, err := NewInvoiceStockRevertRejected(request, ReasonStockAlreadyUsed, shortages)

		return InvoiceRevertDecision{Event: event}, err
	}

	event, err := NewInvoiceStockReverted(request)

	return InvoiceRevertDecision{Movements: movements, Event: event}, err
}

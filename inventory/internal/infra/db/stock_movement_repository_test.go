package db_test

import (
	"testing"

	"inventory/internal/infra/db"
	"inventory/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func storedStock(t *testing.T, productID int) int {
	t.Helper()

	var product model.Product
	require.NoError(t, testConnection.First(&product, productID).Error)

	return product.Stock
}

func TestCreateMovementAddsAConfirmedEntryToTheStock(t *testing.T) {
	movements, products := newMovementRepository(t)

	productID, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})
	require.NoError(t, err)

	id, err := movements.CreateMovement(model.StockMovement{
		ProductID: productID,
		Type:      model.MovementIn,
		Origin:    model.MovementOriginAdjustment,
		Quantity:  10,
		Confirmed: true,
	})
	require.NoError(t, err)

	assert.NotZero(t, id)
	assert.Equal(t, 10, storedStock(t, productID))
}

func TestCreateMovementSubtractsAConfirmedExit(t *testing.T) {
	movements, products := newMovementRepository(t)

	productID, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})
	require.NoError(t, err)

	_, err = movements.CreateMovement(model.StockMovement{
		ProductID: productID, Type: model.MovementIn, Quantity: 10, Confirmed: true,
	})
	require.NoError(t, err)

	_, err = movements.CreateMovement(model.StockMovement{
		ProductID: productID, Type: model.MovementOut, Quantity: 4, Confirmed: true,
	})
	require.NoError(t, err)

	assert.Equal(t, 6, storedStock(t, productID))
}

func TestCreateMovementLeavesTheStockAloneWhenItIsNotConfirmed(t *testing.T) {
	movements, products := newMovementRepository(t)

	productID, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})
	require.NoError(t, err)

	_, err = movements.CreateMovement(model.StockMovement{
		ProductID: productID, Type: model.MovementIn, Quantity: 10, Confirmed: true,
	})
	require.NoError(t, err)

	_, err = movements.CreateMovement(model.StockMovement{
		ProductID: productID, Type: model.MovementOut, Quantity: 4, Confirmed: false,
	})
	require.NoError(t, err)

	assert.Equal(t, 10, storedStock(t, productID), "a reservation does not leave the shelf")
}

func TestCreateMovementLeavesTheOtherProductsAlone(t *testing.T) {
	movements, products := newMovementRepository(t)

	first, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta", Stock: 7})
	require.NoError(t, err)

	second, err := products.CreateProduct(model.Product{Code: "PRD-0002", Name: "Caneca"})
	require.NoError(t, err)

	_, err = movements.CreateMovement(model.StockMovement{
		ProductID: second, Type: model.MovementIn, Quantity: 3, Confirmed: true,
	})
	require.NoError(t, err)

	assert.Equal(t, 7, storedStock(t, first))
	assert.Equal(t, 3, storedStock(t, second))
}

func TestUpdateMovementRecalculatesTheStock(t *testing.T) {
	movements, products := newMovementRepository(t)

	productID, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})
	require.NoError(t, err)

	id, err := movements.CreateMovement(model.StockMovement{
		ProductID: productID, Type: model.MovementIn, Quantity: 10, Confirmed: true,
	})
	require.NoError(t, err)

	err = movements.UpdateMovement(model.StockMovement{
		ID: id, ProductID: productID, Type: model.MovementIn, Quantity: 4, Confirmed: true,
	})
	require.NoError(t, err)

	assert.Equal(t, 4, storedStock(t, productID))
}

func TestUpdateMovementUnconfirmingTakesItOutOfTheStock(t *testing.T) {
	movements, products := newMovementRepository(t)

	productID, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})
	require.NoError(t, err)

	id, err := movements.CreateMovement(model.StockMovement{
		ProductID: productID, Type: model.MovementIn, Quantity: 10, Confirmed: true,
	})
	require.NoError(t, err)

	err = movements.UpdateMovement(model.StockMovement{
		ID: id, ProductID: productID, Type: model.MovementIn, Quantity: 10, Confirmed: false,
	})
	require.NoError(t, err)

	assert.Zero(t, storedStock(t, productID))
}

func TestUpdateMovementWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	movements, _ := newMovementRepository(t)

	err := movements.UpdateMovement(model.StockMovement{
		ID: 9999, ProductID: 1, Type: model.MovementIn, Quantity: 1,
	})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestGetMovementsByProductIDReturnsTheNewestFirst(t *testing.T) {
	movements, products := newMovementRepository(t)

	productID, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})
	require.NoError(t, err)

	other, err := products.CreateProduct(model.Product{Code: "PRD-0002", Name: "Caneca"})
	require.NoError(t, err)

	for _, quantity := range []int{1, 2, 3} {
		_, err = movements.CreateMovement(model.StockMovement{
			ProductID: productID, Type: model.MovementIn, Quantity: quantity, Confirmed: true,
		})
		require.NoError(t, err)
	}

	_, err = movements.CreateMovement(model.StockMovement{
		ProductID: other, Type: model.MovementIn, Quantity: 9, Confirmed: true,
	})
	require.NoError(t, err)

	found, err := movements.GetMovementsByProductID(productID)
	require.NoError(t, err)

	require.Len(t, found, 3, "only the movements of the asked product")
	assert.Equal(t, []int{3, 2, 1}, []int{found[0].Quantity, found[1].Quantity, found[2].Quantity})
}

func TestGetMovementsByProductIDWithoutAnyReturnsAnEmptySlice(t *testing.T) {
	movements, products := newMovementRepository(t)

	productID, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})
	require.NoError(t, err)

	found, err := movements.GetMovementsByProductID(productID)

	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestGetMovementByIDReturnsTheStoredOne(t *testing.T) {
	movements, products := newMovementRepository(t)

	productID, err := products.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})
	require.NoError(t, err)

	invoiceItemID := 3
	id, err := movements.CreateMovement(model.StockMovement{
		ProductID:            productID,
		Type:                 model.MovementOut,
		Origin:               model.MovementOriginInvoice,
		Quantity:             2,
		BillingInvoiceItemID: &invoiceItemID,
	})
	require.NoError(t, err)

	found, err := movements.GetMovementByID(id)

	require.NoError(t, err)
	assert.Equal(t, model.MovementOriginInvoice, found.Origin)
	require.NotNil(t, found.BillingInvoiceItemID)
	assert.Equal(t, 3, *found.BillingInvoiceItemID)
}

func TestGetMovementByIDWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	movements, _ := newMovementRepository(t)

	_, err := movements.GetMovementByID(9999)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestStockMovementKeepsTheInvoiceItCameFrom(t *testing.T) {
	movementRepository, productRepository := newMovementRepository(t)

	productID, err := productRepository.CreateProduct(model.Product{
		Code: "PRD-0001", Name: "Camiseta", Unit: "UN", Price: 30.99,
	})
	require.NoError(t, err)

	invoiceItemID := 108
	billingInvoiceID := 42

	id, err := movementRepository.CreateMovement(model.StockMovement{
		ProductID:            productID,
		Type:                 model.MovementOut,
		Origin:               model.MovementOriginInvoice,
		Quantity:             3,
		Confirmed:            true,
		BillingInvoiceItemID: &invoiceItemID,
		BillingInvoiceID:     &billingInvoiceID,
		InvoiceNumber:        "NF-0042",
	})
	require.NoError(t, err)

	stored, err := movementRepository.GetMovementByID(id)
	require.NoError(t, err)

	require.NotNil(t, stored.BillingInvoiceID)
	assert.Equal(t, 42, *stored.BillingInvoiceID)
	assert.Equal(t, "NF-0042", stored.InvoiceNumber,
		"the number is a snapshot, so the ledger reads on its own")
}

func TestAManualMovementCarriesNoInvoice(t *testing.T) {
	movementRepository, productRepository := newMovementRepository(t)

	productID, err := productRepository.CreateProduct(model.Product{
		Code: "PRD-0001", Name: "Camiseta", Unit: "UN", Price: 30.99,
	})
	require.NoError(t, err)

	id, err := movementRepository.CreateMovement(model.StockMovement{
		ProductID: productID,
		Type:      model.MovementIn,
		Origin:    model.MovementOriginAdjustment,
		Quantity:  5,
		Confirmed: true,
	})
	require.NoError(t, err)

	stored, err := movementRepository.GetMovementByID(id)
	require.NoError(t, err)

	assert.Nil(t, stored.BillingInvoiceID)
	assert.Empty(t, stored.InvoiceNumber)
}

func seedProduct(
	t *testing.T,
	products *db.ProductRepository,
	movements *db.StockMovementRepository,
	code string,
	stock int,
) int {
	t.Helper()

	productID, err := products.CreateProduct(model.Product{
		Code: code, Name: "Parafuso", Unit: "UN", Price: 2.50,
	})
	require.NoError(t, err)

	if stock == 0 {
		return productID
	}

	_, err = movements.CreateMovement(model.StockMovement{
		ProductID: productID,
		Type:      model.MovementIn,
		Origin:    model.MovementOriginAdjustment,
		Quantity:  stock,
		Confirmed: true,
	})
	require.NoError(t, err)

	return productID
}

func stockOf(t *testing.T, products *db.ProductRepository, productID int) int {
	t.Helper()

	product, err := products.GetProductByID(productID)
	require.NoError(t, err)

	return product.Stock
}

func outInvoice(productID int, quantities ...int) model.InvoiceStockRequest {
	items := make([]model.InvoiceStockItem, 0, len(quantities))

	for index, quantity := range quantities {
		items = append(items, model.InvoiceStockItem{
			BillingInvoiceItemID: 100 + index,
			ProductID:            productID,
			Quantity:             quantity,
		})
	}

	return model.InvoiceStockRequest{
		InvoiceID:     42,
		InvoiceNumber: "NF-0042",
		Type:          model.InvoiceTypeOut,
		CausationID:   "cause-1",
		Items:         items,
	}
}

func TestApplyInvoiceTakesTheStockAndRecordsTheResult(t *testing.T) {
	movements, products := newMovementRepository(t)
	productID := seedProduct(t, products, movements, "PROD-1", 20)

	event, err := movements.ApplyInvoice(outInvoice(productID, 8))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockAppliedKey, event.RoutingKey)
	assert.Equal(t, "cause-1", event.CausationID)
	assert.Equal(t, 12, stockOf(t, products, productID))

	var stored []model.StockMovement
	require.NoError(t, testConnection.Where("billing_invoice_id = ?", 42).Find(&stored).Error)

	require.Len(t, stored, 1)
	assert.Equal(t, model.MovementOut, stored[0].Type)
	assert.Equal(t, model.MovementOriginInvoice, stored[0].Origin)
	assert.Equal(t, "NF-0042", stored[0].InvoiceNumber)
	assert.Equal(t, "cause-1", stored[0].CloseEventID)
}

func TestApplyInvoiceAddsTheStockOfAnIncomingInvoice(t *testing.T) {
	movements, products := newMovementRepository(t)
	productID := seedProduct(t, products, movements, "PROD-1", 0)

	request := outInvoice(productID, 7)
	request.Type = model.InvoiceTypeIn

	event, err := movements.ApplyInvoice(request)
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockAppliedKey, event.RoutingKey,
		"an inbound invoice is never refused for balance")
	assert.Equal(t, 7, stockOf(t, products, productID))
}

func TestApplyInvoiceWritesNothingWhenTheStockRefuses(t *testing.T) {
	movements, products := newMovementRepository(t)
	productID := seedProduct(t, products, movements, "PROD-1", 12)

	event, err := movements.ApplyInvoice(outInvoice(productID, 10, 5))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRejectedKey, event.RoutingKey)
	assert.Equal(t, 12, stockOf(t, products, productID), "the balance is untouched")

	var stored []model.StockMovement
	require.NoError(t, testConnection.Where("billing_invoice_id = ?", 42).Find(&stored).Error)
	assert.Empty(t, stored, "the whole transaction rolled back to a refusal")
}

func TestApplyInvoiceTwiceTakesTheStockOnlyOnce(t *testing.T) {
	movements, products := newMovementRepository(t)
	productID := seedProduct(t, products, movements, "PROD-1", 20)

	first, err := movements.ApplyInvoice(outInvoice(productID, 8))
	require.NoError(t, err)

	second, err := movements.ApplyInvoice(outInvoice(productID, 8))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockAppliedKey, first.RoutingKey)
	assert.Equal(t, model.InvoiceStockAppliedKey, second.RoutingKey,
		"a redelivery answers success again, so the invoice does not hang")
	assert.Equal(t, 12, stockOf(t, products, productID), "the second delivery took nothing")

	var stored []model.StockMovement
	require.NoError(t, testConnection.Where("billing_invoice_id = ?", 42).Find(&stored).Error)
	assert.Len(t, stored, 1)
}

func revertOf(invoiceID int) model.InvoiceStockRevertRequest {
	return model.InvoiceStockRevertRequest{
		InvoiceID:     invoiceID,
		InvoiceNumber: "NF-0042",
		CausationID:   "cause-2",
	}
}

func TestRevertInvoiceErasesTheMovementsAndGivesTheStockBack(t *testing.T) {
	movements, products := newMovementRepository(t)
	productID := seedProduct(t, products, movements, "PROD-1", 20)

	_, err := movements.ApplyInvoice(outInvoice(productID, 8))
	require.NoError(t, err)
	require.Equal(t, 12, stockOf(t, products, productID))

	event, err := movements.RevertInvoice(revertOf(42))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRevertedKey, event.RoutingKey)
	assert.Equal(t, 20, stockOf(t, products, productID), "the balance is back where it was")

	var stored []model.StockMovement
	require.NoError(t, testConnection.Where("billing_invoice_id = ?", 42).Find(&stored).Error)
	assert.Empty(t, stored, "the ledger keeps no trace of an invoice that went back to open")
}

func TestRevertInvoiceLetsTheInvoiceBeClosedAgainWithNewQuantities(t *testing.T) {
	movements, products := newMovementRepository(t)
	productID := seedProduct(t, products, movements, "PROD-1", 20)

	_, err := movements.ApplyInvoice(outInvoice(productID, 5))
	require.NoError(t, err)

	_, err = movements.RevertInvoice(revertOf(42))
	require.NoError(t, err)

	_, err = movements.ApplyInvoice(outInvoice(productID, 12))
	require.NoError(t, err)

	assert.Equal(t, 8, stockOf(t, products, productID),
		"the second close takes the corrected quantity, not the old one")
}

func TestRevertInvoiceTwiceIsSafe(t *testing.T) {
	movements, products := newMovementRepository(t)
	productID := seedProduct(t, products, movements, "PROD-1", 20)

	_, err := movements.ApplyInvoice(outInvoice(productID, 8))
	require.NoError(t, err)

	_, err = movements.RevertInvoice(revertOf(42))
	require.NoError(t, err)

	event, err := movements.RevertInvoice(revertOf(42))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRevertedKey, event.RoutingKey)
	assert.Equal(t, 20, stockOf(t, products, productID), "the second revert gives nothing back twice")
}

func TestRevertInvoiceRefusesWhenTheIncomingStockWasAlreadyUsed(t *testing.T) {
	movements, products := newMovementRepository(t)
	productID := seedProduct(t, products, movements, "PROD-1", 0)

	incoming := outInvoice(productID, 10)
	incoming.Type = model.InvoiceTypeIn

	_, err := movements.ApplyInvoice(incoming)
	require.NoError(t, err)
	require.Equal(t, 10, stockOf(t, products, productID))

	_, err = movements.CreateMovement(model.StockMovement{
		ProductID: productID,
		Type:      model.MovementOut,
		Origin:    model.MovementOriginAdjustment,
		Quantity:  7,
		Confirmed: true,
	})
	require.NoError(t, err)

	event, err := movements.RevertInvoice(revertOf(42))
	require.NoError(t, err)

	assert.Equal(t, model.InvoiceStockRevertRejectedKey, event.RoutingKey)
	assert.Equal(t, 3, stockOf(t, products, productID), "the balance is untouched")

	var stored []model.StockMovement
	require.NoError(t, testConnection.Where("billing_invoice_id = ?", 42).Find(&stored).Error)
	assert.Len(t, stored, 1, "the movements stay, the invoice stays closed")
}

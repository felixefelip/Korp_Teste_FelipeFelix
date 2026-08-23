package db_test

import (
	"testing"

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
		ProductID:     productID,
		Type:          model.MovementOut,
		Origin:        model.MovementOriginInvoice,
		Quantity:      2,
		InvoiceItemID: &invoiceItemID,
	})
	require.NoError(t, err)

	found, err := movements.GetMovementByID(id)

	require.NoError(t, err)
	assert.Equal(t, model.MovementOriginInvoice, found.Origin)
	require.NotNil(t, found.InvoiceItemID)
	assert.Equal(t, 3, *found.InvoiceItemID)
}

func TestGetMovementByIDWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	movements, _ := newMovementRepository(t)

	_, err := movements.GetMovementByID(9999)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestApplyInvoiceRecordsTheResultEvent(t *testing.T) {
	movementRepository, _ := newMovementRepository(t)

	event, err := model.NewInvoiceStockApplied(7)
	require.NoError(t, err)

	request := model.InvoiceStockRequest{
		InvoiceID: 7,
		Type:      model.InvoiceTypeOut,
		Items: []model.InvoiceStockItem{
			{InvoiceItemID: 3, ProductID: 42, Quantity: 10},
		},
	}

	require.NoError(t, movementRepository.ApplyInvoice(request, event))

	var events []model.OutboxEvent
	require.NoError(t, testConnection.Find(&events).Error)

	require.Len(t, events, 1)
	assert.Equal(t, model.InvoiceStockAppliedKey, events[0].RoutingKey)
	assert.Equal(t, 7, events[0].AggregateID)
	assert.Nil(t, events[0].PublishedAt, "it is the relay that publishes")
}

package usecase_test

import (
	"testing"

	"inventory/internal/model"
	"inventory/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeMovementRepository struct {
	movements []model.StockMovement
	movement  model.StockMovement
	newID     int
	err       error

	receivedID       int
	receivedMovement model.StockMovement
	receivedRequest  model.InvoiceStockRequest
	recordedEvent    model.OutboxEvent
	calls            int
}

func (f *fakeMovementRepository) ApplyInvoice(request model.InvoiceStockRequest) (model.OutboxEvent, error) {
	f.calls++
	f.receivedRequest = request
	return f.recordedEvent, f.err
}

func (f *fakeMovementRepository) GetMovementsByProductID(productID int) ([]model.StockMovement, error) {
	f.calls++
	f.receivedID = productID
	return f.movements, f.err
}

func (f *fakeMovementRepository) GetMovementByID(id int) (model.StockMovement, error) {
	f.calls++
	f.receivedID = id
	return f.movement, f.err
}

func (f *fakeMovementRepository) CreateMovement(movement model.StockMovement) (int, error) {
	f.calls++
	f.receivedMovement = movement
	return f.newID, f.err
}

func (f *fakeMovementRepository) UpdateMovement(movement model.StockMovement) error {
	f.calls++
	f.receivedMovement = movement
	return f.err
}

func newMovementUsecase(
	movements model.StockMovementRepository,
	products model.ProductRepository,
) usecase.StockMovementUsecase {
	return usecase.NewStockMovementUsecase(movements, products)
}

func TestGetMovementsReturnsWhatTheRepositoryGave(t *testing.T) {
	stored := []model.StockMovement{{ID: 1, ProductID: 7, Type: model.MovementIn, Quantity: 3}}
	movements := &fakeMovementRepository{movements: stored}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{})

	found, err := movementUsecase.GetMovementsByProductID(7)

	require.NoError(t, err)
	assert.Equal(t, stored, found)
	assert.Equal(t, 7, movements.receivedID)
}

func TestGetMovementsWhenTheProductIsGoneStopsBeforeTheLedger(t *testing.T) {
	movements := &fakeMovementRepository{}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{err: gorm.ErrRecordNotFound})

	_, err := movementUsecase.GetMovementsByProductID(7)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Zero(t, movements.calls)
}

func TestCreateMovementForcesTheAdjustmentOrigin(t *testing.T) {
	invoiceItemID := 3
	movements := &fakeMovementRepository{newID: 12}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{})

	created, err := movementUsecase.CreateMovement(model.StockMovement{
		ProductID:            7,
		Type:                 model.MovementOut,
		Origin:               model.MovementOriginInvoice,
		Quantity:             2,
		BillingInvoiceItemID: &invoiceItemID,
	})

	require.NoError(t, err)
	assert.Equal(t, model.MovementOriginAdjustment, created.Origin, "the API never forges a sale")
	assert.Nil(t, created.BillingInvoiceItemID)
	assert.Equal(t, 12, created.ID)
	assert.Equal(t, model.MovementOriginAdjustment, movements.receivedMovement.Origin)
	assert.Nil(t, movements.receivedMovement.BillingInvoiceItemID)
}

func TestCreateMovementStripsAnyInvoiceTheCallerSends(t *testing.T) {
	invoiceItemID := 3
	billingInvoiceID := 42
	movements := &fakeMovementRepository{newID: 12}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{})

	created, err := movementUsecase.CreateMovement(model.StockMovement{
		ProductID:            7,
		Type:                 model.MovementOut,
		Quantity:             2,
		BillingInvoiceItemID: &invoiceItemID,
		BillingInvoiceID:     &billingInvoiceID,
		InvoiceNumber:        "NF-0042",
	})

	require.NoError(t, err)
	assert.Nil(t, created.BillingInvoiceID, "a manual adjustment never belongs to an invoice")
	assert.Empty(t, created.InvoiceNumber)
	assert.Nil(t, movements.receivedMovement.BillingInvoiceID)
	assert.Empty(t, movements.receivedMovement.InvoiceNumber)
}

func TestCreateMovementWhenTheProductIsGoneStopsBeforeTheLedger(t *testing.T) {
	movements := &fakeMovementRepository{}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{err: gorm.ErrRecordNotFound})

	created, err := movementUsecase.CreateMovement(model.StockMovement{ProductID: 7, Quantity: 2})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Zero(t, created)
	assert.Zero(t, movements.calls)
}

func TestUpdateMovementStripsAnyInvoiceTheCallerSends(t *testing.T) {
	billingInvoiceID := 42
	movements := &fakeMovementRepository{
		movement: model.StockMovement{
			ID:        4,
			ProductID: 7,
			Type:      model.MovementIn,
			Origin:    model.MovementOriginAdjustment,
			Quantity:  5,
		},
	}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{})

	updated, err := movementUsecase.UpdateMovement(model.StockMovement{
		ID:               4,
		Quantity:         9,
		BillingInvoiceID: &billingInvoiceID,
		InvoiceNumber:    "NF-0042",
	})

	require.NoError(t, err)
	assert.Nil(t, updated.BillingInvoiceID)
	assert.Empty(t, updated.InvoiceNumber)
}

func TestUpdateMovementKeepsTheProductAndTheOriginOfTheStoredOne(t *testing.T) {
	movements := &fakeMovementRepository{
		movement: model.StockMovement{
			ID:        4,
			ProductID: 7,
			Type:      model.MovementIn,
			Origin:    model.MovementOriginAdjustment,
			Quantity:  5,
		},
	}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{})

	updated, err := movementUsecase.UpdateMovement(model.StockMovement{
		ID:        4,
		ProductID: 99,
		Type:      model.MovementOut,
		Quantity:  2,
		Confirmed: true,
	})

	require.NoError(t, err)
	assert.Equal(t, 7, updated.ProductID, "an edit cannot move the entry to another product")
	assert.Equal(t, model.MovementOriginAdjustment, updated.Origin)
	assert.Equal(t, model.MovementOut, updated.Type)
	assert.Equal(t, 2, updated.Quantity)
	assert.True(t, updated.Confirmed)
}

func TestUpdateMovementRefusesTheOnesBornFromAnInvoice(t *testing.T) {
	invoiceItemID := 3
	movements := &fakeMovementRepository{
		movement: model.StockMovement{ID: 4, ProductID: 7, BillingInvoiceItemID: &invoiceItemID},
	}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{})

	updated, err := movementUsecase.UpdateMovement(model.StockMovement{ID: 4, Quantity: 2})

	require.ErrorIs(t, err, model.ErrMovementFromInvoice)
	assert.Zero(t, updated)
	assert.Equal(t, 1, movements.calls, "it stops at the read, without writing")
}

func TestUpdateMovementWhenItIsGonePropagatesTheError(t *testing.T) {
	movements := &fakeMovementRepository{err: gorm.ErrRecordNotFound}
	movementUsecase := newMovementUsecase(movements, &fakeRepository{})

	updated, err := movementUsecase.UpdateMovement(model.StockMovement{ID: 4})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Zero(t, updated)
}

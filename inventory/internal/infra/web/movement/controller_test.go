package movement_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"inventory/internal/model"
	"inventory/internal/test/webtest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type movementResponse struct {
	ID            int    `json:"id"`
	ProductID     int    `json:"productId"`
	Type          string `json:"type"`
	Origin        string `json:"origin"`
	Quantity      int    `json:"quantity"`
	Confirmed     bool   `json:"confirmed"`
	InvoiceItemID *int   `json:"invoiceItemId"`
}

const validMovement = `{"type":"in","quantity":10,"confirmed":true}`

func decodeMovement(t *testing.T, body []byte) movementResponse {
	t.Helper()

	var movement movementResponse
	require.NoError(t, json.Unmarshal(body, &movement))

	return movement
}

func decodeMovements(t *testing.T, body []byte) []movementResponse {
	t.Helper()

	var movements []movementResponse
	require.NoError(t, json.Unmarshal(body, &movements))

	return movements
}

func createProduct(t *testing.T, server *gin.Engine, body string) int {
	t.Helper()

	response := webtest.Post(t, server, "/products", body)
	require.Equal(t, http.StatusCreated, response.Code)

	var product struct {
		ID int `json:"id"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &product))

	return product.ID
}

func newProduct(t *testing.T) (*gin.Engine, int) {
	t.Helper()

	server := newServer(t)

	return server, createProduct(t, server, `{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)
}

func createMovement(t *testing.T, server *gin.Engine, productID int, body string) int {
	t.Helper()

	response := webtest.Post(t, server, fmt.Sprintf("/products/%d/movements", productID), body)
	require.Equal(t, http.StatusCreated, response.Code)

	return decodeMovement(t, response.Body.Bytes()).ID
}

func productStock(t *testing.T, server *gin.Engine, productID int) int {
	t.Helper()

	response := webtest.Get(t, server, fmt.Sprintf("/products/%d", productID))
	require.Equal(t, http.StatusOK, response.Code)

	var product struct {
		Stock int `json:"stock"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &product))

	return product.Stock
}

func TestCreateMovementReturns201WithTheAdjustmentOrigin(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Post(t, server, fmt.Sprintf("/products/%d/movements", productID), validMovement)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeMovement(t, response.Body.Bytes())

	assert.NotZero(t, created.ID)
	assert.Equal(t, productID, created.ProductID)
	assert.Equal(t, model.MovementIn, created.Type)
	assert.Equal(t, model.MovementOriginAdjustment, created.Origin)
	assert.Equal(t, 10, created.Quantity)
	assert.True(t, created.Confirmed)
	assert.Nil(t, created.InvoiceItemID)
}

func TestCreateMovementUpdatesTheProductStock(t *testing.T) {
	server, productID := newProduct(t)

	createMovement(t, server, productID, validMovement)
	createMovement(t, server, productID, `{"type":"out","quantity":4,"confirmed":true}`)

	assert.Equal(t, 6, productStock(t, server, productID))
}

func TestCreateMovementNotConfirmedLeavesTheStockAlone(t *testing.T) {
	server, productID := newProduct(t)

	createMovement(t, server, productID, validMovement)
	createMovement(t, server, productID, `{"type":"out","quantity":4,"confirmed":false}`)

	assert.Equal(t, 10, productStock(t, server, productID))
}

func TestCreateMovementDefaultsToNotConfirmed(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Post(t, server, fmt.Sprintf("/products/%d/movements", productID),
		`{"type":"in","quantity":5}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.False(t, decodeMovement(t, response.Body.Bytes()).Confirmed)
	assert.Zero(t, productStock(t, server, productID))
}

func TestCreateMovementCannotForgeASale(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Post(t, server, fmt.Sprintf("/products/%d/movements", productID),
		`{"type":"out","quantity":2,"confirmed":true,"origin":"sale","invoiceItemId":7}`)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeMovement(t, response.Body.Bytes())

	assert.Equal(t, model.MovementOriginAdjustment, created.Origin)
	assert.Nil(t, created.InvoiceItemID)
}

func TestCreateMovementForAMissingProductReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products/404/movements", validMovement)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestCreateMovementWithABadProductIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products/abc/movements", validMovement)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCreateMovementWithAnUnknownTypeReturns400WithTheFieldError(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Post(t, server, fmt.Sprintf("/products/%d/movements", productID),
		`{"type":"sideways","quantity":3}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), "type")
}

func TestCreateMovementWithZeroQuantityReturns400(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Post(t, server, fmt.Sprintf("/products/%d/movements", productID),
		`{"type":"in","quantity":0}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), "quantity")
}

func TestCreateMovementWithANegativeQuantityReturns400(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Post(t, server, fmt.Sprintf("/products/%d/movements", productID),
		`{"type":"in","quantity":-3}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCreateMovementWithABrokenBodyReturns400(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Post(t, server, fmt.Sprintf("/products/%d/movements", productID), `{`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestGetMovementsReturnsTheLedgerOfTheProduct(t *testing.T) {
	server, productID := newProduct(t)
	other := createProduct(t, server, `{"code":"PRD-0002","name":"Caneca","unit":"UN","price":5}`)

	createMovement(t, server, productID, validMovement)
	createMovement(t, server, productID, `{"type":"out","quantity":4,"confirmed":true}`)
	createMovement(t, server, other, `{"type":"in","quantity":9,"confirmed":true}`)

	response := webtest.Get(t, server, fmt.Sprintf("/products/%d/movements", productID))

	require.Equal(t, http.StatusOK, response.Code)

	movements := decodeMovements(t, response.Body.Bytes())

	require.Len(t, movements, 2)
	assert.Equal(t, 4, movements[0].Quantity, "the newest comes first")
	assert.Equal(t, 10, movements[1].Quantity)
}

func TestGetMovementsWithoutAnyReturnsAnEmptyArray(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Get(t, server, fmt.Sprintf("/products/%d/movements", productID))

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[]`, response.Body.String(), "an empty ledger is an array, never null")
}

func TestGetMovementsOfAMissingProductReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/products/404/movements")

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestGetMovementByIDReturnsTheMovement(t *testing.T) {
	server, productID := newProduct(t)
	id := createMovement(t, server, productID, validMovement)

	response := webtest.Get(t, server, fmt.Sprintf("/products/%d/movements/%d", productID, id))

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 10, decodeMovement(t, response.Body.Bytes()).Quantity)
}

func TestGetMovementByIDOfAnotherProductReturns404(t *testing.T) {
	server, productID := newProduct(t)
	other := createProduct(t, server, `{"code":"PRD-0002","name":"Caneca","unit":"UN","price":5}`)
	id := createMovement(t, server, productID, validMovement)

	response := webtest.Get(t, server, fmt.Sprintf("/products/%d/movements/%d", other, id))

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestGetMovementByIDWhenMissingReturns404(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Get(t, server, fmt.Sprintf("/products/%d/movements/9999", productID))

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestUpdateMovementReturns200AndRecalculatesTheStock(t *testing.T) {
	server, productID := newProduct(t)
	id := createMovement(t, server, productID, validMovement)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d/movements/%d", productID, id),
		`{"type":"in","quantity":4,"confirmed":true}`)

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeMovement(t, response.Body.Bytes())

	assert.Equal(t, 4, updated.Quantity)
	assert.Equal(t, productID, updated.ProductID)
	assert.Equal(t, model.MovementOriginAdjustment, updated.Origin)
	assert.Equal(t, 4, productStock(t, server, productID))
}

func TestUpdateMovementUnconfirmingTakesItOutOfTheStock(t *testing.T) {
	server, productID := newProduct(t)
	id := createMovement(t, server, productID, validMovement)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d/movements/%d", productID, id),
		`{"type":"in","quantity":10,"confirmed":false}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Zero(t, productStock(t, server, productID))
}

func TestUpdateMovementConfirmingPutsItIntoTheStock(t *testing.T) {
	server, productID := newProduct(t)
	id := createMovement(t, server, productID, `{"type":"in","quantity":7}`)

	require.Zero(t, productStock(t, server, productID))

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d/movements/%d", productID, id),
		`{"type":"in","quantity":7,"confirmed":true}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 7, productStock(t, server, productID))
}

func TestUpdateMovementBornFromAnInvoiceReturns409(t *testing.T) {
	server, productID := newProduct(t)

	invoiceItemID := 3
	stored := model.StockMovement{
		ProductID:     productID,
		Type:          model.MovementOut,
		Origin:        model.MovementOriginSale,
		Quantity:      2,
		Confirmed:     true,
		InvoiceItemID: &invoiceItemID,
	}
	require.NoError(t, testConnection.Create(&stored).Error)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d/movements/%d", productID, stored.ID),
		`{"type":"in","quantity":99,"confirmed":true}`)

	require.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "notas fiscais")
}

func TestUpdateMovementOfAnotherProductReturns404(t *testing.T) {
	server, productID := newProduct(t)
	other := createProduct(t, server, `{"code":"PRD-0002","name":"Caneca","unit":"UN","price":5}`)
	id := createMovement(t, server, productID, validMovement)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d/movements/%d", other, id),
		`{"type":"in","quantity":4,"confirmed":true}`)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestUpdateMovementWhenMissingReturns404(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d/movements/9999", productID),
		`{"type":"in","quantity":4,"confirmed":true}`)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestUpdateMovementWithZeroQuantityReturns400(t *testing.T) {
	server, productID := newProduct(t)
	id := createMovement(t, server, productID, validMovement)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d/movements/%d", productID, id),
		`{"type":"in","quantity":0}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), "quantity")
}

func TestUpdateMovementWithABadIDReturns400(t *testing.T) {
	server, productID := newProduct(t)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d/movements/abc", productID),
		`{"type":"in","quantity":4}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

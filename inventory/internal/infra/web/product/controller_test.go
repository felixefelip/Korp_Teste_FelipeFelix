package product_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"inventory/internal/model"
	"inventory/internal/test/webtest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// productResponse mirrors the public contract of the API, not the model. The
// tests decode into it on purpose: if a new column enters the table without
// entering the response, these tests keep talking only about what the API
// promises.
type productResponse struct {
	ID    int     `json:"id"`
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Unit  string  `json:"unit"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

const validProduct = `{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":12}`

func decodeProduct(t *testing.T, body []byte) productResponse {
	t.Helper()

	var product productResponse
	require.NoError(t, json.Unmarshal(body, &product))

	return product
}

func decodeErrors(t *testing.T, body []byte) map[string]string {
	t.Helper()

	var payload struct {
		Errors map[string]string `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))

	return payload.Errors
}

func TestCreateProductReturns201WithTheCreatedProduct(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products", validProduct)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeProduct(t, response.Body.Bytes())

	assert.NotZero(t, created.ID, "the response should carry the generated id")
	assert.Equal(t, "Camiseta", created.Name)
	assert.Equal(t, 30.99, created.Price)
	assert.Equal(t, 12, created.Stock)
}

func TestCreateProductWithoutStockReturnsZero(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products", `{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Zero(t, decodeProduct(t, response.Body.Bytes()).Stock)
}

func TestCreateProductPersistsToTheDatabase(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products", validProduct)
	require.Equal(t, http.StatusCreated, response.Code)

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, "Camiseta", saved[0].Name)
	assert.Equal(t, 30.99, saved[0].Price)
	assert.Equal(t, 12, saved[0].Stock)
}

func TestCreateProductWithInvalidJSONReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products", `{"name":`)

	assert.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"], "broken JSON has no field to blame, so it becomes a message")

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "nothing should have been stored")
}

func TestCreateProductWithWrongPriceTypeReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":"muito caro"}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["price"])
}

func TestCreateProductWithWrongStockTypeReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":"muitos"}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["stock"])
}

func TestCreateProductWithoutTheRequiredFieldsReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products", `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())
	assert.Equal(t, "Campo obrigatório.", fieldErrors["code"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["name"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["unit"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["price"])
	assert.NotContains(t, fieldErrors, "stock", "a missing stock is valid and becomes zero")

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "an invalid request must not store anything")
}

// Zero price and stock are legitimate: the frontend form starts with stock 0.
// This is the validator trap, which treats a zero value as absent.
func TestCreateProductWithZeroPriceAndStockReturns201(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Amostra gratis","unit":"UN","price":0,"stock":0}`)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeProduct(t, response.Body.Bytes())
	assert.Zero(t, created.Price)
	assert.Zero(t, created.Stock)
}

func TestCreateProductWithNegativePriceReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":-10}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["price"])
}

func TestCreateProductWithNegativeStockReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":-5}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["stock"])
}

func TestCreateProductWithUnitOutsideTheListReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"BANANA","price":30.99}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["unit"])
}

func TestCreateProductWithTextOverTheLimitReturns400(t *testing.T) {
	server := newServer(t)

	longCode := ""
	for range 31 {
		longCode += "X"
	}

	response := webtest.Post(t, server, "/products",
		fmt.Sprintf(`{"code":%q,"name":"Camiseta","unit":"UN","price":30.99}`, longCode))

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["code"],
		"the varchar(30) limit must become a 400, not a database error")
}

// The id belongs to the server: the DTO has no such field, so whatever the
// client sends is ignored instead of becoming the record id.
func TestCreateProductIgnoresTheIDSentByTheClient(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products",
		`{"id":999,"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.NotEqual(t, 999, decodeProduct(t, response.Body.Bytes()).ID)
}

func TestCreateProductNormalizesCodeAndName(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products",
		`{"code":"  prd-0001 ","name":"  Camiseta  ","unit":"UN","price":30.99}`)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeProduct(t, response.Body.Bytes())
	assert.Equal(t, "PRD-0001", created.Code)
	assert.Equal(t, "Camiseta", created.Name)

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)
	require.Len(t, saved, 1)
	assert.Equal(t, "PRD-0001", saved[0].Code, "the database keeps the normalized value")
}

// Contract guard: the response exposes exactly these fields. A new column in
// the model must not leak to the API without going through productResponse.
func TestResponseExposesOnlyTheContractFields(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products", validProduct)
	require.Equal(t, http.StatusCreated, response.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))

	fields := make([]string, 0, len(body))
	for field := range body {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	assert.Equal(t, []string{"code", "id", "name", "price", "stock", "unit"}, fields)
}

func TestGetProductsReturns200WithTheProducts(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/products", validProduct).Code)
	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/products",
		`{"code":"PRD-0002","name":"Calca Jeans","unit":"UN","price":89.99,"stock":3}`).Code)

	response := webtest.Get(t, server, "/products")

	require.Equal(t, http.StatusOK, response.Code)

	var products []productResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &products))

	require.Len(t, products, 2)
	assert.Equal(t, "Camiseta", products[0].Name)
	assert.Equal(t, 30.99, products[0].Price)
	assert.Equal(t, 12, products[0].Stock)
	assert.Equal(t, "Calca Jeans", products[1].Name)
	assert.Equal(t, 89.99, products[1].Price)
	assert.Equal(t, 3, products[1].Stock)
}

func TestGetProductsWhenThereAreNoProducts(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/products")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[]`, response.Body.String())
}

func TestGetProductsWhenTheDatabaseFailsReturns500(t *testing.T) {
	server := newServerWithDatabaseDown(t)

	response := webtest.Get(t, server, "/products")

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var body any
	assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &body),
		"the body must be a single valid JSON, not two concatenated")
}

func TestGetProductByIDReturns200WithTheProduct(t *testing.T) {
	server := newServer(t)

	created := webtest.Post(t, server, "/products", validProduct)
	require.Equal(t, http.StatusCreated, created.Code)

	expected := decodeProduct(t, created.Body.Bytes())

	response := webtest.Get(t, server, fmt.Sprintf("/products/%d", expected.ID))

	require.Equal(t, http.StatusOK, response.Code)

	product := decodeProduct(t, response.Body.Bytes())

	assert.Equal(t, expected.ID, product.ID)
	assert.Equal(t, "Camiseta", product.Name)
	assert.Equal(t, 30.99, product.Price)
	assert.Equal(t, 12, product.Stock)
}

func TestGetProductByIDWhenItDoesNotExistReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/products/404")

	require.Equal(t, http.StatusNotFound, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestGetProductByIDWithNonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/products/abc")

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestCreateProductReturnsCodeAndUnit(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/products", validProduct)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeProduct(t, response.Body.Bytes())
	assert.Equal(t, "PRD-0001", created.Code)
	assert.Equal(t, "UN", created.Unit)
}

func TestCreateProductPersistsCodeAndUnit(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"CX","price":30.99}`).Code)

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, "PRD-0001", saved[0].Code)
	assert.Equal(t, "CX", saved[0].Unit)
}

func TestGetProductsReturnsCodeAndUnit(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/products", validProduct).Code)

	response := webtest.Get(t, server, "/products")
	require.Equal(t, http.StatusOK, response.Code)

	var products []productResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &products))

	require.Len(t, products, 1)
	assert.Equal(t, "PRD-0001", products[0].Code)
	assert.Equal(t, "UN", products[0].Unit)
}

func TestCreateProductAcceptsDuplicateCode(t *testing.T) {
	server := newServer(t)

	first := webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)
	second := webtest.Post(t, server, "/products",
		`{"code":"PRD-0001","name":"Calca Jeans","unit":"UN","price":89.99}`)

	assert.Equal(t, http.StatusCreated, first.Code)
	assert.Equal(t, http.StatusCreated, second.Code, "a duplicate code is allowed")

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Len(t, saved, 2)
}

const updatedProduct = `{"code":"PRD-0002","name":"Camiseta polo","unit":"CX","price":59.9}`

func createProduct(t *testing.T, server *gin.Engine, body string) int {
	t.Helper()

	response := webtest.Post(t, server, "/products", body)
	require.Equal(t, http.StatusCreated, response.Code)

	return decodeProduct(t, response.Body.Bytes()).ID
}

func TestUpdateProductReturns200WithTheUpdatedProduct(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id), updatedProduct)

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeProduct(t, response.Body.Bytes())

	assert.Equal(t, id, updated.ID, "the id must survive the update")
	assert.Equal(t, "PRD-0002", updated.Code)
	assert.Equal(t, "Camiseta polo", updated.Name)
	assert.Equal(t, "CX", updated.Unit)
	assert.Equal(t, 59.9, updated.Price)
	assert.Equal(t, 12, updated.Stock, "the stock is the ledger's, the update does not touch it")
}

func TestUpdateProductPersistsToTheDatabase(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	require.Equal(t, http.StatusOK, webtest.Put(t, server, fmt.Sprintf("/products/%d", id), updatedProduct).Code)

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1, "the update must not create a second row")
	assert.Equal(t, "Camiseta polo", saved[0].Name)
	assert.Equal(t, 59.9, saved[0].Price)
	assert.Equal(t, 12, saved[0].Stock)
}

func TestUpdateProductNormalizesCodeAndName(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id),
		`{"code":"  prd-0009  ","name":"  Caneca  ","unit":"UN","price":10,"stock":1}`)

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeProduct(t, response.Body.Bytes())

	assert.Equal(t, "PRD-0009", updated.Code)
	assert.Equal(t, "Caneca", updated.Name)
}

func TestUpdateProductWithZeroPriceReturns200(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id),
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":0}`)

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeProduct(t, response.Body.Bytes())

	assert.Zero(t, updated.Price)
}

func TestUpdateProductIgnoresTheStockSentInTheBody(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id),
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":999}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 12, decodeProduct(t, response.Body.Bytes()).Stock,
		"only a stock movement moves the balance")
}

func TestUpdateProductIgnoresTheIDSentInTheBody(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id),
		fmt.Sprintf(`{"id":%d,"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":12}`, id+900))

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, id, decodeProduct(t, response.Body.Bytes()).ID, "the id comes from the URL")

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)
	require.Len(t, saved, 1)
}

func TestUpdateProductWhenItDoesNotExistReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Put(t, server, "/products/9999", updatedProduct)

	require.Equal(t, http.StatusNotFound, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "a missing product must not be created by the update")
}

func TestUpdateProductWithNonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Put(t, server, "/products/abc", updatedProduct)

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestUpdateProductWithoutTheRequiredFieldsReturns400(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id), `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())

	assert.Equal(t, "Campo obrigatório.", fieldErrors["code"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["name"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["unit"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["price"])

	var saved model.Product
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Equal(t, "Camiseta", saved.Name, "a rejected update must leave the product untouched")
}

func TestUpdateProductWithNegativePriceReturns400(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id),
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":-1,"stock":12}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "O valor não pode ser negativo.", decodeErrors(t, response.Body.Bytes())["price"])
}

func TestUpdateProductWithUnitOutsideTheListReturns400(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id),
		`{"code":"PRD-0001","name":"Camiseta","unit":"DZ","price":30.99,"stock":12}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["unit"])
}

func TestUpdateProductWithInvalidJSONReturns400(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Put(t, server, fmt.Sprintf("/products/%d", id), `{"code":`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestUpdateProductWithTheDatabaseDownReturns500(t *testing.T) {
	server := newServerWithDatabaseDown(t)

	response := webtest.Put(t, server, "/products/1", updatedProduct)

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestCreateProductRecordsTheInitialStockAsAMovement(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	var movements []model.StockMovement
	require.NoError(t, testConnection.Where("product_id = ?", id).Find(&movements).Error)

	require.Len(t, movements, 1)
	assert.Equal(t, model.MovementIn, movements[0].Type)
	assert.Equal(t, model.MovementOriginAdjustment, movements[0].Origin)
	assert.Equal(t, 12, movements[0].Quantity)
	assert.True(t, movements[0].Confirmed)
	assert.Nil(t, movements[0].InvoiceItemID)
}

func TestCreateProductWithoutStockRecordsNoMovement(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, `{"code":"PRD-0003","name":"Caneca","unit":"UN","price":10}`)

	var movements []model.StockMovement
	require.NoError(t, testConnection.Where("product_id = ?", id).Find(&movements).Error)

	assert.Empty(t, movements)
}

func TestDeleteProductReturns204AndRemovesIt(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	response := webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/products/%d", id), "")

	require.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String(), "204 carries no body")

	assert.Equal(t, http.StatusNotFound,
		webtest.Get(t, server, fmt.Sprintf("/products/%d", id)).Code)
}

func TestDeleteProductRemovesItsMovements(t *testing.T) {
	server := newServer(t)
	id := createProduct(t, server, validProduct)

	require.Equal(t, http.StatusNoContent,
		webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/products/%d", id), "").Code)

	var movements []model.StockMovement
	require.NoError(t, testConnection.Find(&movements).Error)
	assert.Empty(t, movements)
}

func TestDeleteProductTakesItOutOfTheListing(t *testing.T) {
	server := newServer(t)
	first := createProduct(t, server, validProduct)
	createProduct(t, server, `{"code":"PRD-0002","name":"Caneca","unit":"UN","price":10}`)

	require.Equal(t, http.StatusNoContent,
		webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/products/%d", first), "").Code)

	response := webtest.Get(t, server, "/products")

	require.Equal(t, http.StatusOK, response.Code)

	var products []productResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &products))

	require.Len(t, products, 1)
	assert.Equal(t, "PRD-0002", products[0].Code)
}

func TestDeleteProductWhenMissingReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Do(t, server, http.MethodDelete, "/products/9999", "")

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestDeleteProductWithANonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Do(t, server, http.MethodDelete, "/products/abc", "")

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

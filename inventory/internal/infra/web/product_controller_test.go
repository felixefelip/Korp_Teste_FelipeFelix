package web_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"inventory/internal/model"

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

	response := post(t, server, "/products", validProduct)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeProduct(t, response.Body.Bytes())

	assert.NotZero(t, created.ID, "the response should carry the generated id")
	assert.Equal(t, "Camiseta", created.Name)
	assert.Equal(t, 30.99, created.Price)
	assert.Equal(t, 12, created.Stock)
}

func TestCreateProductWithoutStockReturnsZero(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", `{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Zero(t, decodeProduct(t, response.Body.Bytes()).Stock)
}

func TestCreateProductPersistsToTheDatabase(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", validProduct)
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

	response := post(t, server, "/products", `{"name":`)

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

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":"muito caro"}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "tipo invalido", decodeErrors(t, response.Body.Bytes())["price"])
}

func TestCreateProductWithWrongStockTypeReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":"muitos"}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "tipo invalido", decodeErrors(t, response.Body.Bytes())["stock"])
}

func TestCreateProductWithoutTheRequiredFieldsReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())
	assert.Equal(t, "obrigatorio", fieldErrors["code"])
	assert.Equal(t, "obrigatorio", fieldErrors["name"])
	assert.Equal(t, "obrigatorio", fieldErrors["unit"])
	assert.Equal(t, "obrigatorio", fieldErrors["price"])
	assert.NotContains(t, fieldErrors, "stock", "a missing stock is valid and becomes zero")

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "an invalid request must not store anything")
}

// Zero price and stock are legitimate: the frontend form starts with stock 0.
// This is the validator trap, which treats a zero value as absent.
func TestCreateProductWithZeroPriceAndStockReturns201(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Amostra gratis","unit":"UN","price":0,"stock":0}`)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeProduct(t, response.Body.Bytes())
	assert.Zero(t, created.Price)
	assert.Zero(t, created.Stock)
}

func TestCreateProductWithNegativePriceReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":-10}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["price"])
}

func TestCreateProductWithNegativeStockReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":-5}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["stock"])
}

func TestCreateProductWithUnitOutsideTheListReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
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

	response := post(t, server, "/products",
		fmt.Sprintf(`{"code":%q,"name":"Camiseta","unit":"UN","price":30.99}`, longCode))

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["code"],
		"the varchar(30) limit must become a 400, not a database error")
}

// The id belongs to the server: the DTO has no such field, so whatever the
// client sends is ignored instead of becoming the record id.
func TestCreateProductIgnoresTheIDSentByTheClient(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"id":999,"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.NotEqual(t, 999, decodeProduct(t, response.Body.Bytes()).ID)
}

func TestCreateProductNormalizesCodeAndName(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
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

	response := post(t, server, "/products", validProduct)
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

	require.Equal(t, http.StatusCreated, post(t, server, "/products", validProduct).Code)
	require.Equal(t, http.StatusCreated, post(t, server, "/products",
		`{"code":"PRD-0002","name":"Calca Jeans","unit":"UN","price":89.99,"stock":3}`).Code)

	response := get(t, server, "/products")

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

	response := get(t, server, "/products")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[]`, response.Body.String())
}

func TestGetProductsWhenTheDatabaseFailsReturns500(t *testing.T) {
	server := newServerWithDatabaseDown(t)

	response := get(t, server, "/products")

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var body any
	assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &body),
		"the body must be a single valid JSON, not two concatenated")
}

func TestGetProductByIDReturns200WithTheProduct(t *testing.T) {
	server := newServer(t)

	created := post(t, server, "/products", validProduct)
	require.Equal(t, http.StatusCreated, created.Code)

	expected := decodeProduct(t, created.Body.Bytes())

	response := get(t, server, fmt.Sprintf("/products/%d", expected.ID))

	require.Equal(t, http.StatusOK, response.Code)

	product := decodeProduct(t, response.Body.Bytes())

	assert.Equal(t, expected.ID, product.ID)
	assert.Equal(t, "Camiseta", product.Name)
	assert.Equal(t, 30.99, product.Price)
	assert.Equal(t, 12, product.Stock)
}

func TestGetProductByIDWhenItDoesNotExistReturns404(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/products/404")

	require.Equal(t, http.StatusNotFound, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestGetProductByIDWithNonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/products/abc")

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestCreateProductReturnsCodeAndUnit(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", validProduct)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeProduct(t, response.Body.Bytes())
	assert.Equal(t, "PRD-0001", created.Code)
	assert.Equal(t, "UN", created.Unit)
}

func TestCreateProductPersistsCodeAndUnit(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"CX","price":30.99}`).Code)

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, "PRD-0001", saved[0].Code)
	assert.Equal(t, "CX", saved[0].Unit)
}

func TestGetProductsReturnsCodeAndUnit(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, post(t, server, "/products", validProduct).Code)

	response := get(t, server, "/products")
	require.Equal(t, http.StatusOK, response.Code)

	var products []productResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &products))

	require.Len(t, products, 1)
	assert.Equal(t, "PRD-0001", products[0].Code)
	assert.Equal(t, "UN", products[0].Unit)
}

func TestCreateProductAcceptsDuplicateCode(t *testing.T) {
	server := newServer(t)

	first := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)
	second := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Calca Jeans","unit":"UN","price":89.99}`)

	assert.Equal(t, http.StatusCreated, first.Code)
	assert.Equal(t, http.StatusCreated, second.Code, "a duplicate code is allowed")

	var saved []model.Product
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Len(t, saved, 2)
}

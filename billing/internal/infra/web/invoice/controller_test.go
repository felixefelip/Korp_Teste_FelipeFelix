package invoice_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"billing/internal/model"
	"billing/internal/test/webtest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

type invoiceResponse struct {
	ID     int                   `json:"id"`
	Number string                `json:"number"`
	Type   string                `json:"type"`
	Status string                `json:"status"`
	Items  []invoiceItemResponse `json:"items"`
	Total  float64               `json:"total"`
}

const validInvoice = `{"number":"NF-0001","type":"OUT","status":"OPEN"}`

func decodeInvoice(t *testing.T, body []byte) invoiceResponse {
	t.Helper()

	var invoice invoiceResponse
	require.NoError(t, json.Unmarshal(body, &invoice))

	return invoice
}

func decodeInvoices(t *testing.T, body []byte) []invoiceResponse {
	t.Helper()

	var invoices []invoiceResponse
	require.NoError(t, json.Unmarshal(body, &invoices))

	return invoices
}

func seedInvoices(t *testing.T, invoices ...model.Invoice) {
	t.Helper()

	require.NoError(t, testConnection.Create(&invoices).Error)
}

func decodeErrors(t *testing.T, body []byte) map[string]string {
	t.Helper()

	var payload struct {
		Errors map[string]string `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))

	return payload.Errors
}

func TestGetInvoicesReturns200WithEverythingStored(t *testing.T) {
	server := newServer(t)
	seedInvoices(t,
		model.Invoice{Number: "NF-0001", Status: "OPEN"},
		model.Invoice{Number: "NF-0002", Status: "CLOSED"},
	)

	response := webtest.Get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)

	invoices := decodeInvoices(t, response.Body.Bytes())

	require.Len(t, invoices, 2)
	assert.ElementsMatch(t,
		[]string{"NF-0001", "NF-0002"},
		[]string{invoices[0].Number, invoices[1].Number},
	)
}

func TestGetInvoicesCarriesTheFieldsOfEachInvoice(t *testing.T) {
	server := newServer(t)
	seedInvoices(t, model.Invoice{Number: "NF-0001", Status: "CLOSED"})

	response := webtest.Get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)

	invoices := decodeInvoices(t, response.Body.Bytes())

	require.Len(t, invoices, 1)
	assert.NotZero(t, invoices[0].ID)
	assert.Equal(t, "NF-0001", invoices[0].Number)
	assert.Equal(t, "CLOSED", invoices[0].Status)
}

func TestGetInvoicesWithNothingStoredReturnsAnEmptyArray(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[]`, response.Body.String(), "an empty listing must not become null")
}

func TestGetInvoicesListsWhatWasJustCreated(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices", validInvoice).Code)

	response := webtest.Get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)

	invoices := decodeInvoices(t, response.Body.Bytes())

	require.Len(t, invoices, 1)
	assert.Equal(t, "NF-0001", invoices[0].Number)
}

func TestGetInvoicesWithTheDatabaseDownReturns500(t *testing.T) {
	server := newServerWithDatabaseDown(t)

	response := webtest.Get(t, server, "/invoices")

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestGetInvoiceByIDReturns200WithTheInvoice(t *testing.T) {
	server := newServer(t)

	created := webtest.Post(t, server, "/invoices", validInvoice)
	require.Equal(t, http.StatusCreated, created.Code)

	expected := decodeInvoice(t, created.Body.Bytes())

	response := webtest.Get(t, server, fmt.Sprintf("/invoices/%d", expected.ID))

	require.Equal(t, http.StatusOK, response.Code)

	invoice := decodeInvoice(t, response.Body.Bytes())

	assert.Equal(t, expected.ID, invoice.ID)
	assert.Equal(t, "NF-0001", invoice.Number)
	assert.Equal(t, "OPEN", invoice.Status)
}

func TestGetInvoiceByIDWhenItDoesNotExistReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/invoices/404")

	require.Equal(t, http.StatusNotFound, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestGetInvoiceByIDWithNonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/invoices/abc")

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestCreateInvoiceReturns201WithTheCreatedInvoice(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", validInvoice)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeInvoice(t, response.Body.Bytes())

	assert.NotZero(t, created.ID, "the response should carry the generated id")
	assert.Equal(t, "NF-0001", created.Number)
	assert.Equal(t, "OPEN", created.Status)
}

func TestCreateInvoicePersistsToTheDatabase(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", validInvoice)
	require.Equal(t, http.StatusCreated, response.Code)

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, "NF-0001", saved[0].Number)
	assert.Equal(t, "OPEN", saved[0].Status)
}

func TestCreateInvoiceAcceptsTheClosedStatus(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":"NF-0002","type":"OUT","status":"CLOSED"}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "CLOSED", decodeInvoice(t, response.Body.Bytes()).Status)
}

func TestCreateInvoiceUppercasesTheNumber(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":"  nf-0003  ","type":"OUT","status":"OPEN"}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "NF-0003", decodeInvoice(t, response.Body.Bytes()).Number)
}

func TestCreateInvoiceWithoutTheRequiredFieldsReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())

	assert.Equal(t, "Campo obrigatório.", fieldErrors["number"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["type"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["status"])

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "nothing should have been stored")
}

func TestCreateInvoiceWithAnUnknownStatusReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":"NF-0001","type":"OUT","status":"CANCELADA"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["status"])
}

func TestCreateInvoiceWithALowercaseStatusReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":"NF-0001","type":"OUT","status":"open"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["status"])
}

func TestCreateInvoiceWithALongNumberReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices",
		`{"number":"NF-000000000000000000000000000001","type":"OUT","status":"OPEN"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Limite de 30 caracteres excedido.", decodeErrors(t, response.Body.Bytes())["number"])
}

func TestCreateInvoiceWithWrongNumberTypeReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":1,"type":"OUT","status":"OPEN"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["number"])
}

func TestCreateInvoiceWithInvalidJSONReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"], "broken JSON has no field to blame, so it becomes a message")

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "nothing should have been stored")
}

func TestCreateInvoiceWithTheDatabaseDownReturns500(t *testing.T) {
	server := newServerWithDatabaseDown(t)

	response := webtest.Post(t, server, "/invoices", validInvoice)

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

const updatedInvoice = `{"number":"NF-0002","type":"OUT","status":"CLOSED"}`

func createInvoice(t *testing.T, server *gin.Engine, body string) int {
	t.Helper()

	response := webtest.Post(t, server, "/invoices", body)
	require.Equal(t, http.StatusCreated, response.Code)

	return decodeInvoice(t, response.Body.Bytes()).ID
}

func TestUpdateInvoiceReturns200WithTheUpdatedInvoice(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), updatedInvoice)

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeInvoice(t, response.Body.Bytes())

	assert.Equal(t, id, updated.ID, "the id must survive the update")
	assert.Equal(t, "NF-0002", updated.Number)
	assert.Equal(t, "CLOSED", updated.Status)
}

func TestUpdateInvoicePersistsToTheDatabase(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusOK, webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), updatedInvoice).Code)

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1, "the update must not create a second row")
	assert.Equal(t, "NF-0002", saved[0].Number)
	assert.Equal(t, "CLOSED", saved[0].Status)
}

func TestUpdateInvoiceUppercasesTheNumber(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), `{"number":"  nf-0009  ","type":"OUT","status":"OPEN"}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "NF-0009", decodeInvoice(t, response.Body.Bytes()).Number)
}

func TestUpdateInvoiceIgnoresTheIDSentInTheBody(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		fmt.Sprintf(`{"id":%d,"number":"NF-0002","type":"OUT","status":"CLOSED"}`, id+900))

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, id, decodeInvoice(t, response.Body.Bytes()).ID, "the id comes from the URL")

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)
	require.Len(t, saved, 1)
}

func TestUpdateInvoiceWhenItDoesNotExistReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Put(t, server, "/invoices/9999", updatedInvoice)

	require.Equal(t, http.StatusNotFound, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "a missing invoice must not be created by the update")
}

func TestUpdateInvoiceWithNonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Put(t, server, "/invoices/abc", updatedInvoice)

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestUpdateInvoiceWithoutTheRequiredFieldsReturns400(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())

	assert.Equal(t, "Campo obrigatório.", fieldErrors["number"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["status"])
	assert.NotContains(t, fieldErrors, "type", "the update does not take a type at all")

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Equal(t, "NF-0001", saved.Number, "a rejected update must leave the invoice untouched")
}

func TestUpdateInvoiceWithAnUnknownStatusReturns400(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), `{"number":"NF-0001","type":"OUT","status":"CANCELADA"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["status"])
}

func TestUpdateInvoiceWithALongNumberReturns400(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		`{"number":"NF-000000000000000000000000000001","type":"OUT","status":"OPEN"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Limite de 30 caracteres excedido.", decodeErrors(t, response.Body.Bytes())["number"])
}

func TestUpdateInvoiceWithInvalidJSONReturns400(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), `{"number":`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestUpdateInvoiceWithTheDatabaseDownReturns500(t *testing.T) {
	server := newServerWithDatabaseDown(t)

	response := webtest.Put(t, server, "/invoices/1", updatedInvoice)

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestUnknownRouteReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/notas-fiscais")

	assert.Equal(t, http.StatusNotFound, response.Code)
}

const invoiceWithItems = `{
    "number": "NF-0001",
    "type": "OUT",
    "status": "OPEN",
    "items": [
        {"inventoryId": 11, "code": "PRD-0001", "name": "Camiseta", "unit": "UN", "quantity": 2, "unitPrice": 30.99},
        {"inventoryId": 12, "code": "PRD-0002", "name": "Caneca", "unit": "UN", "quantity": 1, "unitPrice": 19.9}
    ]
}`

func itemBody(item string) string {
	return `{"number":"NF-0001","type":"OUT","status":"OPEN","items":[` + item + `]}`
}

func TestCreateInvoiceReturns201WithTheItems(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", invoiceWithItems)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeInvoice(t, response.Body.Bytes())

	require.Len(t, created.Items, 2)
	assert.NotZero(t, created.Items[0].ID)
	assert.Equal(t, "PRD-0001", created.Items[0].Code)
	assert.Equal(t, "Camiseta", created.Items[0].Name)
	assert.Equal(t, "UN", created.Items[0].Unit)
	assert.Equal(t, 2, created.Items[0].Quantity)
	assert.Equal(t, 30.99, created.Items[0].UnitPrice)
	assert.Equal(t, "Caneca", created.Items[1].Name)
}

func TestCreateInvoiceCarriesBothIDsOfTheProduct(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", invoiceWithItems)
	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeInvoice(t, response.Body.Bytes())

	require.Len(t, created.Items, 2)
	assert.Equal(t, 11, created.Items[0].InventoryID, "the id the product has in the inventory")
	assert.NotZero(t, created.Items[0].ProductID, "the id the product has here")
	assert.NotEqual(t, created.Items[0].ProductID, created.Items[1].ProductID)
}

func TestCreateInvoiceTotalsTheItems(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", invoiceWithItems)
	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeInvoice(t, response.Body.Bytes())

	require.Len(t, created.Items, 2)
	assert.Equal(t, 61.98, created.Items[0].Total)
	assert.Equal(t, 19.9, created.Items[1].Total)
	assert.Equal(t, 81.88, created.Total)
}

func TestCreateInvoicePersistsTheItems(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices", invoiceWithItems).Code)

	var items []model.InvoiceItem
	require.NoError(t, testConnection.Find(&items).Error)
	assert.Len(t, items, 2)

	var products []model.Product
	require.NoError(t, testConnection.Find(&products).Error)
	assert.Len(t, products, 2, "each product of the invoice is registered here")
}

func TestCreateInvoiceUppercasesTheCodeOfTheItem(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices",
		itemBody(`{"inventoryId":11,"code":"  prd-0009  ","name":"  Caneca  ","unit":"un","quantity":1,"unitPrice":10}`))

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeInvoice(t, response.Body.Bytes())

	require.Len(t, created.Items, 1)
	assert.Equal(t, "PRD-0009", created.Items[0].Code)
	assert.Equal(t, "Caneca", created.Items[0].Name)
	assert.Equal(t, "UN", created.Items[0].Unit)
}

func TestCreateInvoiceWithoutItemsReturnsAnEmptyArray(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", validInvoice)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeInvoice(t, response.Body.Bytes())

	assert.Empty(t, created.Items)
	assert.Zero(t, created.Total)
	assert.Contains(t, response.Body.String(), `"items":[]`, "an invoice with no items must not become null")
}

func TestCreateInvoiceAcceptsAnItemGivenAway(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices",
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":0}`))

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Zero(t, decodeInvoice(t, response.Body.Bytes()).Items[0].UnitPrice)
}

func TestCreateInvoiceWithAnItemWithoutTheRequiredFieldsReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", itemBody(`{}`))

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())

	assert.Equal(t, "Campo obrigatório.", fieldErrors["items[0].inventoryId"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["items[0].code"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["items[0].name"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["items[0].unit"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["items[0].quantity"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["items[0].unitPrice"])

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "nothing should have been stored")
}

func TestCreateInvoicePointsAtTheItemThatWasRejected(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":"NF-0001","type":"OUT","status":"OPEN","items":[
        {"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":10},
        {"inventoryId":12,"code":"PRD-0002","name":"Caneca","unit":"UN","quantity":0,"unitPrice":10}
    ]}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())

	assert.Equal(t, "O valor precisa ser maior que zero.", fieldErrors["items[1].quantity"])
	assert.NotContains(t, fieldErrors, "items[0].quantity", "the item that is fine must not be blamed")
}

func TestCreateInvoiceWithANegativeQuantityReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server,
		"/invoices",
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":-1,"unitPrice":10}`))

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "O valor precisa ser maior que zero.",
		decodeErrors(t, response.Body.Bytes())["items[0].quantity"])
}

func TestCreateInvoiceWithANegativeUnitPriceReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server,
		"/invoices",
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":-1}`))

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "O valor não pode ser negativo.",
		decodeErrors(t, response.Body.Bytes())["items[0].unitPrice"])
}

func TestCreateInvoiceWithAnItemWithoutTheProductReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server,
		"/invoices",
		itemBody(`{"inventoryId":0,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":10}`))

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["items[0].inventoryId"])
}

func TestGetInvoicesCarriesTheItems(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices", invoiceWithItems).Code)

	response := webtest.Get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)

	invoices := decodeInvoices(t, response.Body.Bytes())

	require.Len(t, invoices, 1)
	require.Len(t, invoices[0].Items, 2)
	assert.Equal(t, "Camiseta", invoices[0].Items[0].Name)
	assert.Equal(t, 81.88, invoices[0].Total)
}

func TestGetInvoiceByIDCarriesTheItems(t *testing.T) {
	server := newServer(t)

	id := createInvoice(t, server, invoiceWithItems)

	response := webtest.Get(t, server, fmt.Sprintf("/invoices/%d", id))

	require.Equal(t, http.StatusOK, response.Code)

	invoice := decodeInvoice(t, response.Body.Bytes())

	require.Len(t, invoice.Items, 2)
	assert.Equal(t, 11, invoice.Items[0].InventoryID)
	assert.Equal(t, 61.98, invoice.Items[0].Total)
}

func TestUpdateInvoiceReplacesTheItems(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		itemBody(`{"inventoryId":13,"code":"PRD-0003","name":"Mochila","unit":"UN","quantity":3,"unitPrice":99.9}`))

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeInvoice(t, response.Body.Bytes())

	require.Len(t, updated.Items, 1)
	assert.Equal(t, "Mochila", updated.Items[0].Name)
	assert.Equal(t, 299.7, updated.Total)

	var items []model.InvoiceItem
	require.NoError(t, testConnection.Find(&items).Error)
	assert.Len(t, items, 1, "the replaced items must not stay behind")
}

func TestUpdateInvoiceWithAnEmptyListEmptiesTheInvoice(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		`{"number":"NF-0001","type":"OUT","status":"OPEN","items":[]}`)

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeInvoice(t, response.Body.Bytes())

	assert.Empty(t, updated.Items)
	assert.Zero(t, updated.Total)
}

func TestUpdateInvoiceWithoutTheItemsKeyKeepsTheItems(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), updatedInvoice)

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeInvoice(t, response.Body.Bytes())

	assert.Equal(t, "NF-0002", updated.Number)
	require.Len(t, updated.Items, 2, "an update that says nothing about the items must not erase them")
	assert.Equal(t, 81.88, updated.Total)
}

func TestUpdateInvoiceWithARejectedItemKeepsTheInvoiceUntouched(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		itemBody(`{"inventoryId":13,"code":"PRD-0003","name":"Mochila","unit":"UN","quantity":0,"unitPrice":99.9}`))

	require.Equal(t, http.StatusBadRequest, response.Code)

	invoice := decodeInvoice(t, webtest.Get(t, server, fmt.Sprintf("/invoices/%d", id)).Body.Bytes())

	require.Len(t, invoice.Items, 2, "a rejected update must leave the items untouched")
	assert.Equal(t, "Camiseta", invoice.Items[0].Name)
}

func TestUpdateInvoiceKeepsSellingTheSameProduct(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":5,"unitPrice":30.99}`))

	require.Equal(t, http.StatusOK, response.Code)

	var products []model.Product
	require.NoError(t, testConnection.Find(&products).Error)
	assert.Len(t, products, 2, "the products already registered are reused")
}

func TestCreateInvoiceCarriesTheTypeBack(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", validInvoice)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, model.InvoiceTypeOut, decodeInvoice(t, response.Body.Bytes()).Type)
}

func TestCreateInvoiceAcceptsAnInboundInvoice(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":"NF-0004","type":"IN","status":"OPEN"}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, model.InvoiceTypeIn, decodeInvoice(t, response.Body.Bytes()).Type)

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, model.InvoiceTypeIn, saved[0].Type)
}

func TestCreateInvoiceWithAnUnknownTypeReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":"NF-0001","type":"ENTRADA","status":"OPEN"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["type"])
}

func TestCreateInvoiceRejectsALowercaseType(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"number":"NF-0001","type":"out","status":"OPEN"}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestUpdateInvoiceIgnoresTheTypeSentInTheBody(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		`{"number":"NF-0001","type":"IN","status":"OPEN"}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, model.InvoiceTypeOut, decodeInvoice(t, response.Body.Bytes()).Type,
		"the direction is settled at issue")
}

func TestUpdateInvoiceWithoutATypeIsAccepted(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, `{"number":"NF-0009","type":"IN","status":"OPEN"}`)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		`{"number":"NF-0009","status":"CLOSED"}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, model.InvoiceTypeIn, decodeInvoice(t, response.Body.Bytes()).Type)
}

func replicaOf(t *testing.T, inventoryID int) model.Product {
	t.Helper()

	var product model.Product
	require.NoError(t, testConnection.Where("inventory_id = ?", inventoryID).First(&product).Error)

	return product
}

func TestOutboundInvoiceRewritesThePriceOfTheReplica(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices",
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":30.99}`)).Code)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices",
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":45.5}`)).Code)

	assert.Equal(t, 45.5, replicaOf(t, 11).Price, "a sale price is the catalogue price")
}

func TestInboundInvoiceKeepsThePriceOfTheReplica(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices",
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":30.99}`)).Code)

	inbound := `{"number":"NF-0009","type":"IN","status":"OPEN","items":[
        {"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":10,"unitPrice":12.5}
    ]}`
	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices", inbound).Code)

	assert.Equal(t, 30.99, replicaOf(t, 11).Price,
		"a purchase price must not become the catalogue price")
}

func TestInboundInvoiceStillRefreshesTheRestOfTheReplica(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices",
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":30.99}`)).Code)

	inbound := `{"number":"NF-0009","type":"IN","status":"OPEN","items":[
        {"inventoryId":11,"code":"PRD-0001","name":"Camiseta polo","unit":"CX","quantity":10,"unitPrice":12.5}
    ]}`
	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices", inbound).Code)

	replica := replicaOf(t, 11)

	assert.Equal(t, "Camiseta polo", replica.Name)
	assert.Equal(t, "CX", replica.Unit)
}

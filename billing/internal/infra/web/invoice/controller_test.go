package invoice_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"billing/internal/infra/db"
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
	ID              int                   `json:"id"`
	Series          int                   `json:"series"`
	Number          int                   `json:"number"`
	FormattedNumber string                `json:"formattedNumber"`
	Type            string                `json:"type"`
	Status          string                `json:"status"`
	Items           []invoiceItemResponse `json:"items"`
	Total           float64               `json:"total"`
	ProcessingSince *time.Time            `json:"processingSince"`
}

const validInvoice = `{"series":1,"number":1,"type":"OUT","status":"OPEN"}`

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
		model.Invoice{Series: 1, Number: 1, Status: "OPEN"},
		model.Invoice{Series: 1, Number: 2, Status: "CLOSED"},
	)

	response := webtest.Get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)

	invoices := decodeInvoices(t, response.Body.Bytes())

	require.Len(t, invoices, 2)
	assert.ElementsMatch(t,
		[]int{1, 2},
		[]int{invoices[0].Number, invoices[1].Number},
	)
}

func TestGetInvoicesCarriesTheFieldsOfEachInvoice(t *testing.T) {
	server := newServer(t)
	seedInvoices(t, model.Invoice{Series: 1, Number: 1, Status: "CLOSED"})

	response := webtest.Get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)

	invoices := decodeInvoices(t, response.Body.Bytes())

	require.Len(t, invoices, 1)
	assert.NotZero(t, invoices[0].ID)
	assert.Equal(t, 1, invoices[0].Number)
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
	assert.Equal(t, 1, invoices[0].Number)
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
	assert.Equal(t, 1, invoice.Number)
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
	assert.Equal(t, 1, created.Number)
	assert.Equal(t, "OPEN", created.Status)
}

func TestCreateInvoicePersistsToTheDatabase(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", validInvoice)
	require.Equal(t, http.StatusCreated, response.Code)

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, 1, saved[0].Number)
	assert.Equal(t, "OPEN", saved[0].Status)
}

func TestCreateInvoiceFormatsTheDocumentNumber(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"series":1,"number":57,"type":"OUT"}`)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeInvoice(t, response.Body.Bytes())
	assert.Equal(t, 1, created.Series)
	assert.Equal(t, 57, created.Number)
	assert.Equal(t, "001/000057", created.FormattedNumber)
}

func TestCreateInvoiceWithADuplicateDocumentReturns409(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated,
		webtest.Post(t, server, "/invoices", `{"series":1,"number":57,"type":"OUT"}`).Code)

	response := webtest.Post(t, server, "/invoices", `{"series":1,"number":57,"type":"IN"}`)

	require.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "série e número")
}

func TestCreateInvoiceWithoutTheRequiredFieldsReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())

	assert.Equal(t, "Campo obrigatório.", fieldErrors["number"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["type"])
	assert.NotContains(t, fieldErrors, "status", "the status is not the caller's to pick")

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "nothing should have been stored")
}

func TestCreateInvoiceWithANumberOverTheLimitReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices",
		`{"series":1,"number":1000000,"type":"OUT"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "O valor não pode ser maior que 999999.", decodeErrors(t, response.Body.Bytes())["number"])
}

func TestCreateInvoiceWithASeriesOverTheLimitReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"series":1000,"number":1,"type":"OUT"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "O valor não pode ser maior que 999.", decodeErrors(t, response.Body.Bytes())["series"])
}

func TestCreateInvoiceWithWrongNumberTypeReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"series":1,"number":"NF-1","type":"OUT"}`)

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

const updatedInvoice = `{"series":1,"number":2}`

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
	assert.Equal(t, 2, updated.Number)
	assert.Equal(t, "OPEN", updated.Status, "only the close action moves the status")
}

func TestUpdateInvoicePersistsToTheDatabase(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusOK, webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), updatedInvoice).Code)

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1, "the update must not create a second row")
	assert.Equal(t, 2, saved[0].Number)
	assert.Equal(t, "OPEN", saved[0].Status)
}

func TestUpdateInvoiceDocumentNumberIsRewritten(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), `{"series":2,"number":9}`)

	require.Equal(t, http.StatusOK, response.Code)

	updated := decodeInvoice(t, response.Body.Bytes())
	assert.Equal(t, 2, updated.Series)
	assert.Equal(t, 9, updated.Number)
	assert.Equal(t, "002/000009", updated.FormattedNumber)
}

func TestUpdateInvoiceIgnoresTheIDSentInTheBody(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		fmt.Sprintf(`{"id":%d,"series":1,"number":2,"type":"OUT","status":"CLOSED"}`, id+900))

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
	assert.NotContains(t, fieldErrors, "type", "the update does not take a type at all")
	assert.NotContains(t, fieldErrors, "status", "nor a status")

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)
	assert.Equal(t, 1, saved.Number, "a rejected update must leave the invoice untouched")
}

func TestUpdateInvoiceIgnoresTheStatusSentInTheBody(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		`{"series":1,"number":1,"status":"CLOSED"}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "OPEN", decodeInvoice(t, response.Body.Bytes()).Status,
		"closing is an action of its own, not a field of the update")
}

func TestUpdateInvoiceWithANumberOverTheLimitReturns400(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		`{"series":1,"number":1000000}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "O valor não pode ser maior que 999999.", decodeErrors(t, response.Body.Bytes())["number"])
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
    "series":1,"number":1,
    "type": "OUT",
    "status": "OPEN",
    "items": [
        {"inventoryId": 11, "code": "PRD-0001", "name": "Camiseta", "unit": "UN", "quantity": 2, "unitPrice": 30.99},
        {"inventoryId": 12, "code": "PRD-0002", "name": "Caneca", "unit": "UN", "quantity": 1, "unitPrice": 19.9}
    ]
}`

func itemBody(item string) string {
	return `{"series":1,"number":1,"type":"OUT","status":"OPEN","items":[` + item + `]}`
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

	response := webtest.Post(t, server, "/invoices", `{"series":1,"number":1,"type":"OUT","status":"OPEN","items":[
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
		`{"series":1,"number":1,"type":"OUT","status":"OPEN","items":[]}`)

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

	assert.Equal(t, 2, updated.Number)
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

	response := webtest.Post(t, server, "/invoices", `{"series":1,"number":4,"type":"IN","status":"OPEN"}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, model.InvoiceTypeIn, decodeInvoice(t, response.Body.Bytes()).Type)

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, model.InvoiceTypeIn, saved[0].Type)
}

func TestCreateInvoiceWithAnUnknownTypeReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"series":1,"number":1,"type":"ENTRADA","status":"OPEN"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErrors(t, response.Body.Bytes())["type"])
}

func TestCreateInvoiceRejectsALowercaseType(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices", `{"series":1,"number":1,"type":"out","status":"OPEN"}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestUpdateInvoiceIgnoresTheTypeSentInTheBody(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		`{"series":1,"number":1,"type":"IN","status":"OPEN"}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, model.InvoiceTypeOut, decodeInvoice(t, response.Body.Bytes()).Type,
		"the direction is settled at issue")
}

func TestUpdateInvoiceWithoutATypeIsAccepted(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, `{"series":1,"number":9,"type":"IN"}`)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
		`{"series":1,"number":9,"status":"CLOSED"}`)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, model.InvoiceTypeIn, decodeInvoice(t, response.Body.Bytes()).Type)
}

func replicaOf(t *testing.T, inventoryID int) model.Product {
	t.Helper()

	var product model.Product
	require.NoError(t, testConnection.Where("inventory_id = ?", inventoryID).First(&product).Error)

	return product
}

func TestInboundInvoiceKeepsThePriceOfTheReplica(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices",
		itemBody(`{"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":1,"unitPrice":30.99}`)).Code)

	inbound := `{"series":1,"number":9,"type":"IN","status":"OPEN","items":[
        {"inventoryId":11,"code":"PRD-0001","name":"Camiseta","unit":"UN","quantity":10,"unitPrice":12.5}
    ]}`
	require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices", inbound).Code)

	assert.Equal(t, 30.99, replicaOf(t, 11).Price,
		"a purchase price must not become the catalogue price")
}

func TestDeleteInvoiceReturns204AndRemovesIt(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/invoices/%d", id), "")

	require.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String(), "204 carries no body")

	assert.Equal(t, http.StatusNotFound,
		webtest.Get(t, server, fmt.Sprintf("/invoices/%d", id)).Code)
}

func TestDeleteInvoiceRemovesItsItems(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	require.Equal(t, http.StatusNoContent,
		webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/invoices/%d", id), "").Code)

	var items []model.InvoiceItem
	require.NoError(t, testConnection.Find(&items).Error)
	assert.Empty(t, items)
}

func TestDeleteAClosedInvoiceReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)
	closeInvoiceCompletely(t, server, id)

	response := webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/invoices/%d", id), "")

	require.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "fechadas")
}

func TestDeleteAClosedInvoiceKeepsItStored(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)
	closeInvoiceCompletely(t, server, id)

	require.Equal(t, http.StatusConflict,
		webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/invoices/%d", id), "").Code)

	assert.Equal(t, http.StatusOK, webtest.Get(t, server, fmt.Sprintf("/invoices/%d", id)).Code)
}

func TestDeleteInvoiceWhenMissingReturns404(t *testing.T) {
	server := newServer(t)

	response := webtest.Do(t, server, http.MethodDelete, "/invoices/9999", "")

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestDeleteInvoiceWithANonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Do(t, server, http.MethodDelete, "/invoices/abc", "")

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func closeInvoice(t *testing.T, server *gin.Engine, id int) *httptest.ResponseRecorder {
	t.Helper()

	return webtest.Post(t, server, fmt.Sprintf("/invoices/%d/close", id), "")
}

// closeInvoiceCompletely pede o fechamento e aplica o resultado do inventory,
// que em producao chegaria pela fila billing.stock-results.
func closeInvoiceCompletely(t *testing.T, server *gin.Engine, id int) {
	t.Helper()

	require.Equal(t, http.StatusAccepted, closeInvoice(t, server, id).Code)

	moved, err := db.NewInvoiceRepository(testConnection).ConfirmClose(id)
	require.NoError(t, err)
	require.True(t, moved)
}

func TestCloseInvoiceReturns202WithTheInvoiceBeingProcessed(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := closeInvoice(t, server, id)

	require.Equal(t, http.StatusAccepted, response.Code)

	closing := decodeInvoice(t, response.Body.Bytes())

	assert.Equal(t, id, closing.ID)
	assert.Equal(t, "CLOSING", closing.Status, "the stock has not been taken yet")
	assert.Equal(t, 1, closing.Number, "closing changes the status and nothing else")
}

func TestCloseInvoiceOnlyReachesClosedWhenTheStockIsApplied(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)

	response := webtest.Get(t, server, fmt.Sprintf("/invoices/%d", id))
	require.Equal(t, http.StatusOK, response.Code)

	assert.Equal(t, "CLOSED", decodeInvoice(t, response.Body.Bytes()).Status)
}

func TestCloseInvoicePersistsToTheDatabase(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusAccepted, closeInvoice(t, server, id).Code)

	var saved model.Invoice
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "CLOSING", saved.Status)
}

func TestCloseInvoiceKeepsTheItemsAndTheTotal(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	response := closeInvoice(t, server, id)

	require.Equal(t, http.StatusAccepted, response.Code)

	closing := decodeInvoice(t, response.Body.Bytes())

	assert.Len(t, closing.Items, 2)
	assert.Equal(t, 81.88, closing.Total)
}

func TestCloseAnAlreadyClosedInvoiceReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)

	response := closeInvoice(t, server, id)

	require.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "já está fechada")
}

func TestCloseAnInvoiceBeingProcessedReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusAccepted, closeInvoice(t, server, id).Code)

	response := closeInvoice(t, server, id)

	require.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "em processamento")
}

func TestAnInvoiceBeingProcessedCannotBeEditedOrDeleted(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusAccepted, closeInvoice(t, server, id).Code)

	update := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), validInvoice)
	require.Equal(t, http.StatusConflict, update.Code)
	assert.Contains(t, update.Body.String(), "em processamento")

	remove := webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/invoices/%d", id), "")
	require.Equal(t, http.StatusConflict, remove.Code)
	assert.Contains(t, remove.Body.String(), "em processamento")
}

func TestCloseInvoiceWhenMissingReturns404(t *testing.T) {
	server := newServer(t)

	response := closeInvoice(t, server, 9999)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestCloseInvoiceWithANonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices/abc/close", "")

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestAClosedInvoiceCannotBeDeleted(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)

	assert.Equal(t, http.StatusConflict,
		webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/invoices/%d", id), "").Code)
}

func reopenInvoiceCompletely(t *testing.T, server *gin.Engine, id int) {
	t.Helper()

	require.Equal(t, http.StatusAccepted, reopenInvoice(t, server, id).Code)

	moved, err := db.NewInvoiceRepository(testConnection).ConfirmReopen(id)
	require.NoError(t, err)
	require.True(t, moved)
}

func reopenInvoice(t *testing.T, server *gin.Engine, id int) *httptest.ResponseRecorder {
	t.Helper()

	return webtest.Post(t, server, fmt.Sprintf("/invoices/%d/reopen", id), "")
}

func TestReopenInvoiceReturns202WithTheInvoiceBeingProcessed(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)

	response := reopenInvoice(t, server, id)

	require.Equal(t, http.StatusAccepted, response.Code)

	reopening := decodeInvoice(t, response.Body.Bytes())

	assert.Equal(t, id, reopening.ID)
	assert.Equal(t, "REOPENING", reopening.Status, "the stock has not been given back yet")
	assert.Equal(t, 1, reopening.Number)
}

func TestAnInvoiceBeingReopenedCannotBeEditedOrDeleted(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)
	require.Equal(t, http.StatusAccepted, reopenInvoice(t, server, id).Code)

	update := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), validInvoice)
	require.Equal(t, http.StatusConflict, update.Code)
	assert.Contains(t, update.Body.String(), "em processamento")

	remove := webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/invoices/%d", id), "")
	require.Equal(t, http.StatusConflict, remove.Code)
	assert.Contains(t, remove.Body.String(), "em processamento")
}

func TestReopenAnAlreadyOpenInvoiceReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := reopenInvoice(t, server, id)

	require.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "já está aberta")
}

func TestReopenInvoiceWhenMissingReturns404(t *testing.T) {
	server := newServer(t)

	response := reopenInvoice(t, server, 9999)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestReopenInvoiceWithANonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices/abc/reopen", "")

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestAReopenedInvoiceCanBeDeletedAgain(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)
	reopenInvoiceCompletely(t, server, id)

	assert.Equal(t, http.StatusNoContent,
		webtest.Do(t, server, http.MethodDelete, fmt.Sprintf("/invoices/%d", id), "").Code)
}

func TestCreateInvoiceAlwaysStartsOpen(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices",
		`{"series":1,"number":5,"type":"OUT","status":"CLOSED"}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "OPEN", decodeInvoice(t, response.Body.Bytes()).Status,
		"a brand new invoice is open, whatever the body says")

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, model.InvoiceStatusOpen, saved[0].Status)
}

func TestUpdateAClosedInvoiceReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)

	response := webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), updatedInvoice)

	require.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "alteradas")
}

func TestUpdateAClosedInvoiceKeepsItAsItWas(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	closeInvoiceCompletely(t, server, id)

	require.Equal(t, http.StatusConflict,
		webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id),
			`{"series":1,"number":9999,"items":[]}`).Code)

	invoice := decodeInvoice(t, webtest.Get(t, server, fmt.Sprintf("/invoices/%d", id)).Body.Bytes())

	assert.Equal(t, 1, invoice.Number)
	assert.Len(t, invoice.Items, 2, "a refused update must not have erased the items")
}

func TestAReopenedInvoiceCanBeUpdatedAgain(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)
	reopenInvoiceCompletely(t, server, id)

	assert.Equal(t, http.StatusOK,
		webtest.Put(t, server, fmt.Sprintf("/invoices/%d", id), updatedInvoice).Code)
}

func printDanfe(t *testing.T, server *gin.Engine, id int) *httptest.ResponseRecorder {
	t.Helper()

	return webtest.Get(t, server, fmt.Sprintf("/invoices/%d/danfe", id))
}

func TestPrintDanfeOfAClosedInvoiceReturns200WithAPdf(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, invoiceWithItems)

	closeInvoiceCompletely(t, server, id)

	response := printDanfe(t, server, id)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "application/pdf", response.Header().Get("Content-Type"))
	assert.True(t, bytes.HasPrefix(response.Body.Bytes(), []byte("%PDF-")))
}

func TestPrintDanfeNamesTheFileAfterTheInvoiceNumber(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)

	response := printDanfe(t, server, id)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t,
		`inline; filename="danfe-001-000001.pdf"`,
		response.Header().Get("Content-Disposition"),
		"the slash of the formatted number would break the file name",
	)
}

func TestPrintDanfeOfAnOpenInvoiceReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := printDanfe(t, server, id)

	assert.Equal(t, http.StatusConflict, response.Code, "there is nothing to print yet")
}

func TestPrintDanfeOfAnInvoiceBeingProcessedReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusAccepted, closeInvoice(t, server, id).Code)

	response := printDanfe(t, server, id)

	assert.Equal(t, http.StatusConflict, response.Code, "the stock has not been taken yet")
}

func TestPrintDanfeOfAReopenedInvoiceReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)
	require.Equal(t, http.StatusAccepted,
		webtest.Post(t, server, fmt.Sprintf("/invoices/%d/reopen", id), "").Code)

	moved, err := db.NewInvoiceRepository(testConnection).ConfirmReopen(id)
	require.NoError(t, err)
	require.True(t, moved)

	response := printDanfe(t, server, id)

	assert.Equal(t, http.StatusConflict, response.Code, "reopening takes the danfe away")
}

func TestPrintDanfeOfAnInvoiceThatDoesNotExistReturns404(t *testing.T) {
	server := newServer(t)

	assert.Equal(t, http.StatusNotFound, printDanfe(t, server, 9999).Code)
}

func TestPrintDanfeWithAnIdThatIsNotANumberReturns400(t *testing.T) {
	server := newServer(t)

	assert.Equal(t, http.StatusBadRequest, webtest.Get(t, server, "/invoices/abc/danfe").Code)
}

func retryInvoice(t *testing.T, server *gin.Engine, id int) *httptest.ResponseRecorder {
	t.Helper()

	return webtest.Post(t, server, fmt.Sprintf("/invoices/%d/retry", id), "")
}

func recordedEvents(t *testing.T, invoiceID int, routingKey string) []model.OutboxEvent {
	t.Helper()

	var events []model.OutboxEvent

	err := testConnection.
		Where("aggregate_id = ? AND routing_key = ?", invoiceID, routingKey).
		Order("id").
		Find(&events).Error
	require.NoError(t, err)

	return events
}

func TestRetryInvoiceReturns202AndAsksTheInventoryAgain(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusAccepted, closeInvoice(t, server, id).Code)

	response := retryInvoice(t, server, id)

	require.Equal(t, http.StatusAccepted, response.Code)

	retried := decodeInvoice(t, response.Body.Bytes())

	assert.Equal(t, "CLOSING", retried.Status, "retrying resends the request, it does not settle it")

	events := recordedEvents(t, id, model.InvoiceCloseRequestedKey)

	require.Len(t, events, 2, "the second request is the one the inventory never answered")
	assert.NotEqual(t, events[0].EventID, events[1].EventID)
}

func TestRetryAnInvoiceBeingReopenedAsksForTheRevertAgain(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)
	require.Equal(t, http.StatusAccepted, reopenInvoice(t, server, id).Code)

	require.Equal(t, http.StatusAccepted, retryInvoice(t, server, id).Code)

	assert.Len(t, recordedEvents(t, id, model.InvoiceReopenRequestedKey), 2)
	assert.Len(t, recordedEvents(t, id, model.InvoiceCloseRequestedKey), 1, "the close is not resent")
}

func TestRetryAnInvoiceThatIsNotBeingProcessedReturns409(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	response := retryInvoice(t, server, id)

	require.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "não está em processamento")
	assert.Empty(t, recordedEvents(t, id, model.InvoiceCloseRequestedKey))
}

func TestRetryInvoiceWhenMissingReturns404(t *testing.T) {
	server := newServer(t)

	response := retryInvoice(t, server, 9999)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestRetryInvoiceWithANonNumericIDReturns400(t *testing.T) {
	server := newServer(t)

	response := webtest.Post(t, server, "/invoices/abc/retry", "")

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestAnInvoiceBeingProcessedTellsSinceWhen(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusAccepted, closeInvoice(t, server, id).Code)

	closing := decodeInvoice(t, webtest.Get(t, server, fmt.Sprintf("/invoices/%d", id)).Body.Bytes())

	require.NotNil(t, closing.ProcessingSince, "the screen decides what to say from this")
	assert.WithinDuration(t, time.Now(), *closing.ProcessingSince, time.Minute)
}

func TestASettledInvoiceTellsNoProcessingClock(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	closeInvoiceCompletely(t, server, id)

	closed := decodeInvoice(t, webtest.Get(t, server, fmt.Sprintf("/invoices/%d", id)).Body.Bytes())

	require.Equal(t, "CLOSED", closed.Status)
	assert.Nil(t, closed.ProcessingSince)
}

func TestTheListingCarriesTheProcessingClock(t *testing.T) {
	server := newServer(t)
	id := createInvoice(t, server, validInvoice)

	require.Equal(t, http.StatusAccepted, closeInvoice(t, server, id).Code)

	listed := decodeInvoices(t, webtest.Get(t, server, "/invoices").Body.Bytes())

	require.Len(t, listed, 1)
	assert.NotNil(t, listed[0].ProcessingSince)
}

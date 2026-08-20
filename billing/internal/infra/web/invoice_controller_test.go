package web_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type invoiceResponse struct {
	ID     int    `json:"id"`
	Number string `json:"number"`
	Status string `json:"status"`
}

const validInvoice = `{"number":"NF-0001","status":"OPEN"}`

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

	response := get(t, server, "/invoices")

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

	response := get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)

	invoices := decodeInvoices(t, response.Body.Bytes())

	require.Len(t, invoices, 1)
	assert.NotZero(t, invoices[0].ID)
	assert.Equal(t, "NF-0001", invoices[0].Number)
	assert.Equal(t, "CLOSED", invoices[0].Status)
}

func TestGetInvoicesWithNothingStoredReturnsAnEmptyArray(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[]`, response.Body.String(), "an empty listing must not become null")
}

func TestGetInvoicesListsWhatWasJustCreated(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, post(t, server, "/invoices", validInvoice).Code)

	response := get(t, server, "/invoices")

	require.Equal(t, http.StatusOK, response.Code)

	invoices := decodeInvoices(t, response.Body.Bytes())

	require.Len(t, invoices, 1)
	assert.Equal(t, "NF-0001", invoices[0].Number)
}

func TestGetInvoicesWithTheDatabaseDownReturns500(t *testing.T) {
	server := newServerWithDatabaseDown(t)

	response := get(t, server, "/invoices")

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestCreateInvoiceReturns201WithTheCreatedInvoice(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", validInvoice)

	require.Equal(t, http.StatusCreated, response.Code)

	created := decodeInvoice(t, response.Body.Bytes())

	assert.NotZero(t, created.ID, "the response should carry the generated id")
	assert.Equal(t, "NF-0001", created.Number)
	assert.Equal(t, "OPEN", created.Status)
}

func TestCreateInvoicePersistsToTheDatabase(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", validInvoice)
	require.Equal(t, http.StatusCreated, response.Code)

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)

	require.Len(t, saved, 1)
	assert.Equal(t, "NF-0001", saved[0].Number)
	assert.Equal(t, "OPEN", saved[0].Status)
}

func TestCreateInvoiceAcceptsTheClosedStatus(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", `{"number":"NF-0002","status":"CLOSED"}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "CLOSED", decodeInvoice(t, response.Body.Bytes()).Status)
}

func TestCreateInvoiceUppercasesTheNumber(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", `{"number":"  nf-0003  ","status":"OPEN"}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "NF-0003", decodeInvoice(t, response.Body.Bytes()).Number)
}

func TestCreateInvoiceWithoutTheRequiredFieldsReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	fieldErrors := decodeErrors(t, response.Body.Bytes())

	assert.Equal(t, "Campo obrigatório.", fieldErrors["number"])
	assert.Equal(t, "Campo obrigatório.", fieldErrors["status"])

	var saved []model.Invoice
	require.NoError(t, testConnection.Find(&saved).Error)
	assert.Empty(t, saved, "nothing should have been stored")
}

func TestCreateInvoiceWithAnUnknownStatusReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", `{"number":"NF-0001","status":"CANCELADA"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["status"])
}

func TestCreateInvoiceWithALowercaseStatusReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", `{"number":"NF-0001","status":"open"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["status"])
}

func TestCreateInvoiceWithALongNumberReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices",
		`{"number":"NF-000000000000000000000000000001","status":"OPEN"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Limite de 30 caracteres excedido.", decodeErrors(t, response.Body.Bytes())["number"])
}

func TestCreateInvoiceWithWrongNumberTypeReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", `{"number":1,"status":"OPEN"}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Valor inválido.", decodeErrors(t, response.Body.Bytes())["number"])
}

func TestCreateInvoiceWithInvalidJSONReturns400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/invoices", `{"number":`)

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

	response := post(t, server, "/invoices", validInvoice)

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEmpty(t, body["message"])
}

func TestUnknownRouteReturns404(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/notas-fiscais")

	assert.Equal(t, http.StatusNotFound, response.Code)
}

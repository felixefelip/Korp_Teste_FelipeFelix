package invoice_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"billing/internal/test/webtest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type documentResponse struct {
	Series int  `json:"series"`
	Number *int `json:"number"`
}

func nextDocument(t *testing.T, server *gin.Engine, query string) documentResponse {
	t.Helper()

	response := webtest.Get(t, server, "/invoices/next-document"+query)
	require.Equal(t, http.StatusOK, response.Code)

	var document documentResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &document))

	return document
}

func TestNextDocumentOnAnEmptyDatabase(t *testing.T) {
	server := newServer(t)

	document := nextDocument(t, server, "")

	assert.Equal(t, 1, document.Series)
	require.NotNil(t, document.Number)
	assert.Equal(t, 1, *document.Number)
}

func TestNextDocumentFollowsTheLastInvoiceOfTheSeries(t *testing.T) {
	server := newServer(t)

	for _, number := range []int{1, 2, 7} {
		body := fmt.Sprintf(`{"series":1,"number":%d,"type":"OUT"}`, number)
		require.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices", body).Code)
	}

	document := nextDocument(t, server, "")

	assert.Equal(t, 1, document.Series)
	require.NotNil(t, document.Number)
	assert.Equal(t, 8, *document.Number)
}

func TestNextDocumentCountsEachSeriesOnItsOwn(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated,
		webtest.Post(t, server, "/invoices", `{"series":1,"number":9,"type":"OUT"}`).Code)

	document := nextDocument(t, server, "?series=2")

	assert.Equal(t, 2, document.Series)
	require.NotNil(t, document.Number)
	assert.Equal(t, 1, *document.Number)
}

func TestNextDocumentPicksTheHighestSeriesInUse(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated,
		webtest.Post(t, server, "/invoices", `{"series":1,"number":9,"type":"OUT"}`).Code)
	require.Equal(t, http.StatusCreated,
		webtest.Post(t, server, "/invoices", `{"series":4,"number":3,"type":"OUT"}`).Code)

	document := nextDocument(t, server, "")

	assert.Equal(t, 4, document.Series)
	require.NotNil(t, document.Number)
	assert.Equal(t, 4, *document.Number)
}

func TestNextDocumentSuggestionCanBeSaved(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated,
		webtest.Post(t, server, "/invoices", `{"series":1,"number":1,"type":"OUT"}`).Code)

	document := nextDocument(t, server, "")
	body := fmt.Sprintf(`{"series":%d,"number":%d,"type":"OUT"}`, document.Series, *document.Number)

	assert.Equal(t, http.StatusCreated, webtest.Post(t, server, "/invoices", body).Code)
}

func TestNextDocumentRejectsAnInvalidSeries(t *testing.T) {
	server := newServer(t)

	for _, query := range []string{"?series=abc", "?series=0", "?series=-1"} {
		response := webtest.Get(t, server, "/invoices/next-document"+query)

		assert.Equal(t, http.StatusBadRequest, response.Code, "query %q", query)
	}
}

func TestNextDocumentDoesNotShadowTheInvoiceByID(t *testing.T) {
	server := newServer(t)

	created := webtest.Post(t, server, "/invoices", validInvoice)
	require.Equal(t, http.StatusCreated, created.Code)

	invoice := decodeInvoice(t, created.Body.Bytes())
	response := webtest.Get(t, server, fmt.Sprintf("/invoices/%d", invoice.ID))

	assert.Equal(t, http.StatusOK, response.Code)
}

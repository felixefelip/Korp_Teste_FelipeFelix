package invoice_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"billing/internal/test/webtest"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newServerWithoutExtractor(t *testing.T) *gin.Engine {
	t.Helper()

	t.Setenv("ANTHROPIC_API_KEY", "")

	return newServer(t)
}

func TestDraftInvoiceWithoutPrompt(t *testing.T) {
	server := newServerWithoutExtractor(t)

	response := webtest.Post(t, server, "/invoices/draft", `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	var body struct {
		Errors map[string]string `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "Campo obrigatório.", body.Errors["prompt"])
}

func TestDraftInvoiceWithMalformedBody(t *testing.T) {
	server := newServerWithoutExtractor(t)

	response := webtest.Post(t, server, "/invoices/draft", `{"prompt":`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestDraftInvoiceWhenTheExtractorIsNotConfigured(t *testing.T) {
	server := newServerWithoutExtractor(t)

	response := webtest.Post(t, server, "/invoices/draft", `{"prompt":"vender 2 notebooks"}`)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)

	var body struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Contains(t, body.Message, "não está configurado")
}

func TestDraftInvoiceDoesNotShadowTheInvoiceRoutes(t *testing.T) {
	server := newServerWithoutExtractor(t)

	response := webtest.Post(t, server, "/invoices", validInvoice)

	assert.Equal(t, http.StatusCreated, response.Code)
}

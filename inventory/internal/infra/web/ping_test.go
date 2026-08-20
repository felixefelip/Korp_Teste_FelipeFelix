package web_test

import (
	"net/http"
	"testing"

	"inventory/internal/test/webtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPing covers the healthcheck route, which used to live inline in main.go
// and for that reason was not reachable by a test.
func TestPing(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/ping")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"message":"pong"}`, response.Body.String())
}

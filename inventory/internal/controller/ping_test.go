package controller_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPing covers the healthcheck route, which used to live inline in main.go
// and for that reason was not reachable by a test.
func TestPing(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/ping")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"message":"pong"}`, response.Body.String())
}

package controller_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPing cobre a rota de healthcheck, que antes vivia inline no main.go e
// por isso nao era alcancavel por teste.
func TestPing(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/ping")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"message":"pong"}`, response.Body.String())
}

package web_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/ping")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"message":"pong"}`, response.Body.String())
}

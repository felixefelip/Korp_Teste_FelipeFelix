package web_test

import (
	"net/http"
	"testing"

	"billing/internal/test/webtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	server := newServer(t)

	response := webtest.Get(t, server, "/ping")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"message":"pong"}`, response.Body.String())
}

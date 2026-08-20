// Package webtest holds the HTTP helpers shared by the tests of the web packages.
package webtest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func Do(t testing.TB, server *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	return response
}

func Get(t testing.TB, server *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()

	return Do(t, server, http.MethodGet, path, "")
}

func Post(t testing.TB, server *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	return Do(t, server, http.MethodPost, path, body)
}

func Put(t testing.TB, server *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	return Do(t, server, http.MethodPut, path, body)
}

package web_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"inventory/internal/infra/db"
	"inventory/internal/infra/web"
	"inventory/internal/test/dbtest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// testConnection is the connection shared by every test in the package.
var testConnection *gorm.DB

func TestMain(m *testing.M) {
	connection, err := dbtest.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, "test database setup:", err)
		os.Exit(1)
	}

	testConnection = connection
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}

// newServer builds the server with the real application routes (the same ones
// main.go registers, through web.Register) over an empty product table.
func newServer(t *testing.T) *gin.Engine {
	t.Helper()

	dbtest.Reset(t, testConnection)

	server := gin.New()
	web.Register(server, testConnection)

	return server
}

// newServerWithDatabaseDown builds the same server over a connection of its own
// that has already been closed, to exercise the error path of the handlers. The
// connection shared by the other tests is not affected.
func newServerWithDatabaseDown(t *testing.T) *gin.Engine {
	t.Helper()

	connection, err := db.ConnectDB()
	require.NoError(t, err)

	sqlDB, err := connection.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	server := gin.New()
	web.Register(server, connection)

	return server
}

// do fires a request against the server and returns the recorded response.
func do(t *testing.T, server *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	return response
}

func get(t *testing.T, server *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()

	return do(t, server, http.MethodGet, path, "")
}

func post(t *testing.T, server *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	return do(t, server, http.MethodPost, path, body)
}

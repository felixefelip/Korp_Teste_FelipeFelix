package web_test

import (
	"fmt"
	"os"
	"testing"

	"inventory/internal/infra/web"
	"inventory/internal/test/dbtest"

	"github.com/gin-gonic/gin"
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

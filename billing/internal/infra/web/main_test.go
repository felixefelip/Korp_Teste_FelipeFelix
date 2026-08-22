package web_test

import (
	"fmt"
	"os"
	"testing"

	"billing/internal/infra/web"
	"billing/internal/test/dbtest"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

func newServer(t *testing.T) *gin.Engine {
	t.Helper()

	dbtest.Reset(t, testConnection)

	server := gin.New()
	web.Register(server, testConnection)

	return server
}

package repository

import (
	"fmt"
	"os"
	"testing"

	"estoque/internal/testdb"

	"gorm.io/gorm"
)

var testConnection *gorm.DB

func TestMain(m *testing.M) {
	connection, err := testdb.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, "setup do banco de teste:", err)
		os.Exit(1)
	}

	testConnection = connection

	os.Exit(m.Run())
}

func newRepository(t *testing.T) ProductRepository {
	t.Helper()

	testdb.Reset(t, testConnection)

	return NewProductRepository(testConnection)
}

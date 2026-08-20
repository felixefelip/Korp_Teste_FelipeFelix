package db_test

import (
	"fmt"
	"os"
	"testing"

	"inventory/internal/infra/db"
	"inventory/internal/testdb"

	"gorm.io/gorm"
)

var testConnection *gorm.DB

func TestMain(m *testing.M) {
	connection, err := testdb.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, "test database setup:", err)
		os.Exit(1)
	}

	testConnection = connection

	os.Exit(m.Run())
}

func newRepository(t *testing.T) db.ProductRepository {
	t.Helper()

	testdb.Reset(t, testConnection)

	return db.NewProductRepository(testConnection)
}

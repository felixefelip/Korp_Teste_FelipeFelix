package db_test

import (
	"fmt"
	"os"
	"testing"

	"billing/internal/infra/db"
	"billing/internal/test/dbtest"

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

	os.Exit(m.Run())
}

func newRepository(t *testing.T) *db.InvoiceRepository {
	t.Helper()

	dbtest.Reset(t, testConnection)

	return db.NewInvoiceRepository(testConnection)
}

func newOutboxRepository(t *testing.T) *db.OutboxRepository {
	t.Helper()

	dbtest.Reset(t, testConnection)

	return db.NewOutboxRepository(testConnection)
}

func newProductRepository(t *testing.T) *db.ProductRepository {
	t.Helper()

	dbtest.Reset(t, testConnection)

	return db.NewProductRepository(testConnection)
}

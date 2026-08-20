// Package testdb holds the setup of the database shared by the integration
// tests of the several packages.
package testdb

import (
	"fmt"
	"os"
	"testing"

	"inventory/internal/infra/db"
	"inventory/internal/model"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Setup makes sure the test database exists and is migrated, returning the
// connection. It must be called from the TestMain of each package.
func Setup() (*gorm.DB, error) {
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "test")
	}

	// Safety latch: the tests truncate tables, so they may only run against
	// the test database.
	if db.Env() != "test" {
		return nil, fmt.Errorf("refusing to run with GO_ENV=%q; use GO_ENV=test", db.Env())
	}

	if err := db.EnsureDatabase(); err != nil {
		return nil, fmt.Errorf("creating the database: %w", err)
	}

	connection, err := db.ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("connecting to the database: %w", err)
	}

	if err := connection.AutoMigrate(&model.Product{}); err != nil {
		return nil, fmt.Errorf("migrating the database: %w", err)
	}

	return connection, nil
}

// Reset returns the database to the empty state, so that one test does not see
// the state left behind by another.
func Reset(t testing.TB, connection *gorm.DB) {
	t.Helper()

	err := connection.Exec("TRUNCATE TABLE product RESTART IDENTITY CASCADE").Error
	require.NoError(t, err, "cleaning the product table")
}

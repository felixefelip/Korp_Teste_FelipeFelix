// Package dbtest holds the setup of the database shared by the integration
// tests of the several packages.
package dbtest

import (
	"fmt"
	"os"
	"testing"

	"inventory/internal/infra/db"
	"inventory/internal/model"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Setup() (*gorm.DB, error) {
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "test")
	}

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

	if err := connection.AutoMigrate(&model.Product{}, &model.StockMovement{}); err != nil {
		return nil, fmt.Errorf("migrating the database: %w", err)
	}

	return connection, nil
}

func Reset(t testing.TB, connection *gorm.DB) {
	t.Helper()

	err := connection.Exec("TRUNCATE TABLE product, stock_movement RESTART IDENTITY CASCADE").Error
	require.NoError(t, err, "cleaning the product and stock_movement tables")
}

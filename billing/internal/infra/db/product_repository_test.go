package db_test

import (
	"testing"

	"billing/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func catalogProduct(inventoryID int, code, name string, price float64) model.Product {
	return model.Product{
		InventoryID: inventoryID,
		Code:        code,
		Name:        name,
		Unit:        "UN",
		Price:       price,
	}
}

func TestUpsertProductRegistersWhatTheSyncBrings(t *testing.T) {
	repository := newProductRepository(t)

	require.NoError(t, repository.UpsertProduct(catalogProduct(11, "PRD-0001", "Camiseta", 30.99)))

	products, err := repository.GetProducts()
	require.NoError(t, err)

	require.Len(t, products, 1)
	assert.Equal(t, 11, products[0].InventoryID)
	assert.Equal(t, "Camiseta", products[0].Name)
}

func TestUpsertProductOverwritesTheProductItAlreadyKnew(t *testing.T) {
	repository := newProductRepository(t)

	require.NoError(t, repository.UpsertProduct(catalogProduct(11, "PRD-0001", "Camiseta", 30.99)))
	require.NoError(t, repository.UpsertProduct(catalogProduct(11, "PRD-0001", "Camiseta polo", 45.5)))

	products, err := repository.GetProducts()
	require.NoError(t, err)

	require.Len(t, products, 1, "the same product from the inventory is one row here")
	assert.Equal(t, "Camiseta polo", products[0].Name)
	assert.Equal(t, 45.5, products[0].Price)
}

func TestGetProductsSortsByCode(t *testing.T) {
	repository := newProductRepository(t)

	require.NoError(t, repository.UpsertProduct(catalogProduct(12, "PRD-0002", "Calça", 80)))
	require.NoError(t, repository.UpsertProduct(catalogProduct(11, "PRD-0001", "Camiseta", 30.99)))

	products, err := repository.GetProducts()
	require.NoError(t, err)

	assert.Equal(t, []string{"PRD-0001", "PRD-0002"}, []string{products[0].Code, products[1].Code})
}

func TestDeactivateProductTakesItOutOfTheCatalog(t *testing.T) {
	repository := newProductRepository(t)

	require.NoError(t, repository.UpsertProduct(catalogProduct(11, "PRD-0001", "Camiseta", 30.99)))
	require.NoError(t, repository.DeactivateProduct(11))

	products, err := repository.GetProducts()
	require.NoError(t, err)

	assert.Empty(t, products)
}

func TestDeactivateProductKeepsTheRowSoOldInvoicesStillResolve(t *testing.T) {
	repository := newProductRepository(t)

	require.NoError(t, repository.UpsertProduct(catalogProduct(11, "PRD-0001", "Camiseta", 30.99)))
	require.NoError(t, repository.DeactivateProduct(11))

	var stored model.Product
	require.NoError(t, testConnection.Where("inventory_id = ?", 11).First(&stored).Error)

	assert.False(t, stored.Active)
	assert.Equal(t, "Camiseta", stored.Name, "the invoice items still point at this row")
}

func TestUpsertProductBringsBackAProductThatCameAgain(t *testing.T) {
	repository := newProductRepository(t)

	require.NoError(t, repository.UpsertProduct(catalogProduct(11, "PRD-0001", "Camiseta", 30.99)))
	require.NoError(t, repository.DeactivateProduct(11))
	require.NoError(t, repository.UpsertProduct(catalogProduct(11, "PRD-0001", "Camiseta", 30.99)))

	products, err := repository.GetProducts()
	require.NoError(t, err)

	assert.Len(t, products, 1)
}

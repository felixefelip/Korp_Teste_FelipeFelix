package db_test

import (
	"testing"

	"inventory/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateProduct(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateProduct(model.Product{Name: "Camiseta", Price: 30.99, Stock: 12})
	require.NoError(t, err)
	assert.NotZero(t, id, "the database should have generated an id")

	var saved model.Product
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "Camiseta", saved.Name)
	assert.Equal(t, 30.99, saved.Price)
	assert.Equal(t, 12, saved.Stock)
}

func TestCreateProductWithoutStockStoresZero(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateProduct(model.Product{Name: "Camiseta", Price: 30.99})
	require.NoError(t, err)

	var saved model.Product
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Zero(t, saved.Stock, "with no stock informed the product must stay zeroed")
}

func TestGetProductsWhenEmpty(t *testing.T) {
	repository := newRepository(t)

	products, err := repository.GetProducts()

	require.NoError(t, err)
	assert.Empty(t, products)
}

func TestGetProductsReturnsTheCreatedProducts(t *testing.T) {
	repository := newRepository(t)

	created := []model.Product{
		{Name: "Camiseta", Price: 30.99, Stock: 12},
		{Name: "Calca Jeans", Price: 89.99, Stock: 3},
	}

	for _, product := range created {
		_, err := repository.CreateProduct(product)
		require.NoError(t, err)
	}

	products, err := repository.GetProducts()
	require.NoError(t, err)
	require.Len(t, products, len(created))

	for i, expected := range created {
		assert.Equal(t, expected.Name, products[i].Name)
		assert.Equal(t, expected.Price, products[i].Price)
		assert.Equal(t, expected.Stock, products[i].Stock)
	}
}

func TestIsolationBetweenTests(t *testing.T) {
	repository := newRepository(t)

	products, err := repository.GetProducts()

	require.NoError(t, err)
	assert.Empty(t, products, "state leaked from another test")
}

func TestGetProductByIDReturnsTheProduct(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateProduct(model.Product{Name: "Camiseta", Price: 30.99, Stock: 12})
	require.NoError(t, err)

	product, err := repository.GetProductByID(id)

	require.NoError(t, err)
	assert.Equal(t, id, product.ID)
	assert.Equal(t, "Camiseta", product.Name)
	assert.Equal(t, 30.99, product.Price)
	assert.Equal(t, 12, product.Stock)
}

func TestGetProductByIDWhenMissingReturnsErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.GetProductByID(404)

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCreateProductStoresCodeAndUnit(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateProduct(model.Product{
		Code: "PRD-0001", Name: "Camiseta", Unit: "UN", Price: 30.99, Stock: 12,
	})
	require.NoError(t, err)

	var saved model.Product
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "PRD-0001", saved.Code)
	assert.Equal(t, "UN", saved.Unit)
}

// Business rule: two products may share the same code.
func TestCreateProductAllowsDuplicateCode(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta", Unit: "UN"})
	require.NoError(t, err)

	second, err := repository.CreateProduct(model.Product{Code: "PRD-0001", Name: "Calca Jeans", Unit: "UN"})
	require.NoError(t, err, "a duplicate code must be accepted")

	assert.NotEqual(t, first, second, "each product has its own id")

	products, err := repository.GetProducts()
	require.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestCreateProductWithoutCodeAndUnitStoresEmpty(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateProduct(model.Product{Name: "Camiseta", Price: 30.99})
	require.NoError(t, err)

	var saved model.Product
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Empty(t, saved.Code)
	assert.Empty(t, saved.Unit)
}

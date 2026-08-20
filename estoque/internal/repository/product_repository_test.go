package repository

import (
	"testing"

	"estoque/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateProduct(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateProduct(model.Product{Name: "Camiseta", Price: 30.99, Stock: 12})
	require.NoError(t, err)
	assert.NotZero(t, id, "o banco deveria ter gerado um id")

	var saved model.Product
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Equal(t, "Camiseta", saved.Name)
	assert.Equal(t, 30.99, saved.Price)
	assert.Equal(t, 12, saved.Stock)
}

func TestCreateProductSemStockGravaZero(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateProduct(model.Product{Name: "Camiseta", Price: 30.99})
	require.NoError(t, err)

	var saved model.Product
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Zero(t, saved.Stock, "sem stock informado o produto deve ficar zerado")
}

func TestGetProductsQuandoVazio(t *testing.T) {
	repository := newRepository(t)

	products, err := repository.GetProducts()

	require.NoError(t, err)
	assert.Empty(t, products)
}

func TestGetProductsRetornaOsProdutosCriados(t *testing.T) {
	repository := newRepository(t)

	criados := []model.Product{
		{Name: "Camiseta", Price: 30.99, Stock: 12},
		{Name: "Calca Jeans", Price: 89.99, Stock: 3},
	}

	for _, product := range criados {
		_, err := repository.CreateProduct(product)
		require.NoError(t, err)
	}

	products, err := repository.GetProducts()
	require.NoError(t, err)
	require.Len(t, products, len(criados))

	for i, esperado := range criados {
		assert.Equal(t, esperado.Name, products[i].Name)
		assert.Equal(t, esperado.Price, products[i].Price)
		assert.Equal(t, esperado.Stock, products[i].Stock)
	}
}

func TestIsolamentoEntreTestes(t *testing.T) {
	repository := newRepository(t)

	products, err := repository.GetProducts()

	require.NoError(t, err)
	assert.Empty(t, products, "estado vazou de outro teste")
}

func TestGetProductByIDRetornaOProduto(t *testing.T) {
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

func TestGetProductByIDQuandoNaoExisteRetornaErrRecordNotFound(t *testing.T) {
	repository := newRepository(t)

	_, err := repository.GetProductByID(404)

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCreateProductGravaCodigoEUnidade(t *testing.T) {
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

// Regra de negocio: dois produtos podem compartilhar o mesmo codigo.
func TestCreateProductPermiteCodigoDuplicado(t *testing.T) {
	repository := newRepository(t)

	primeiro, err := repository.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta", Unit: "UN"})
	require.NoError(t, err)

	segundo, err := repository.CreateProduct(model.Product{Code: "PRD-0001", Name: "Calca Jeans", Unit: "UN"})
	require.NoError(t, err, "codigo duplicado deve ser aceito")

	assert.NotEqual(t, primeiro, segundo, "cada produto tem seu proprio id")

	products, err := repository.GetProducts()
	require.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestCreateProductSemCodigoEUnidadeGravaVazio(t *testing.T) {
	repository := newRepository(t)

	id, err := repository.CreateProduct(model.Product{Name: "Camiseta", Price: 30.99})
	require.NoError(t, err)

	var saved model.Product
	require.NoError(t, testConnection.First(&saved, id).Error)

	assert.Empty(t, saved.Code)
	assert.Empty(t, saved.Unit)
}

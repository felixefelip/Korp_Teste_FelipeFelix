package usecase_test

import (
	"errors"
	"testing"

	"inventory/internal/model"
	"inventory/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	products []model.Product
	product  model.Product
	newID    int
	err      error

	receivedID      int
	receivedProduct model.Product
	calls           int
}

func (f *fakeRepository) GetProducts() ([]model.Product, error) {
	f.calls++
	return f.products, f.err
}

func (f *fakeRepository) GetProductByID(id int) (model.Product, error) {
	f.calls++
	f.receivedID = id
	return f.product, f.err
}

func (f *fakeRepository) CreateProduct(product model.Product) (int, error) {
	f.calls++
	f.receivedProduct = product
	return f.newID, f.err
}

func (f *fakeRepository) UpdateProduct(product model.Product) error {
	f.calls++
	f.receivedProduct = product
	return f.err
}

var errRepository = errors.New("database down")

func newUsecase(repository model.ProductRepository) usecase.ProductUsecase {
	return usecase.NewProductUsecase(repository)
}

func TestGetProductsReturnsWhatTheRepositoryGave(t *testing.T) {
	stored := []model.Product{{ID: 1, Code: "PRD-0001", Name: "Camiseta"}}
	repository := &fakeRepository{products: stored}
	productUsecase := newUsecase(repository)

	products, err := productUsecase.GetProducts()

	require.NoError(t, err)
	assert.Equal(t, stored, products)
	assert.Equal(t, 1, repository.calls)
}

func TestGetProductsPropagatesTheRepositoryError(t *testing.T) {
	productUsecase := newUsecase(&fakeRepository{err: errRepository})

	_, err := productUsecase.GetProducts()

	assert.ErrorIs(t, err, errRepository)
}

func TestGetProductByIDForwardsTheID(t *testing.T) {
	repository := &fakeRepository{product: model.Product{ID: 7, Name: "Camiseta"}}
	productUsecase := newUsecase(repository)

	product, err := productUsecase.GetProductByID(7)

	require.NoError(t, err)
	assert.Equal(t, 7, repository.receivedID)
	assert.Equal(t, "Camiseta", product.Name)
}

func TestGetProductByIDPropagatesTheRepositoryError(t *testing.T) {
	productUsecase := newUsecase(&fakeRepository{err: errRepository})

	_, err := productUsecase.GetProductByID(7)

	assert.ErrorIs(t, err, errRepository)
}

func TestCreateProductReturnsTheProductWithTheGeneratedID(t *testing.T) {
	repository := &fakeRepository{newID: 42}
	productUsecase := newUsecase(repository)

	created, err := productUsecase.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})

	require.NoError(t, err)
	assert.Equal(t, 42, created.ID, "the id must come from the repository")
	assert.Equal(t, "Camiseta", created.Name)
}

func TestCreateProductHandsTheProductToTheRepositoryUntouched(t *testing.T) {
	repository := &fakeRepository{newID: 42}
	productUsecase := newUsecase(repository)

	product := model.Product{Code: "PRD-0001", Name: "Camiseta", Unit: "UN", Price: 30.99, Stock: 12}
	_, err := productUsecase.CreateProduct(product)

	require.NoError(t, err)
	assert.Equal(t, product, repository.receivedProduct)
	assert.Zero(t, repository.receivedProduct.ID, "the id is the database's to assign")
}

func TestCreateProductWhenTheRepositoryFailsReturnsTheZeroValue(t *testing.T) {
	productUsecase := newUsecase(&fakeRepository{err: errRepository})

	created, err := productUsecase.CreateProduct(model.Product{Name: "Camiseta"})

	assert.ErrorIs(t, err, errRepository)
	assert.Equal(t, model.Product{}, created, "nothing partially filled leaks out on failure")
}

func TestUpdateProductReturnsTheProductItSaved(t *testing.T) {
	repository := &fakeRepository{}
	productUsecase := newUsecase(repository)

	product := model.Product{ID: 7, Code: "PRD-0007", Name: "Camiseta", Price: 30.99, Stock: 4}
	updated, err := productUsecase.UpdateProduct(product)

	require.NoError(t, err)
	assert.Equal(t, product, updated)
	assert.Equal(t, product, repository.receivedProduct)
	assert.Equal(t, 1, repository.calls)
}

func TestUpdateProductPropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeRepository{err: errRepository}
	productUsecase := newUsecase(repository)

	updated, err := productUsecase.UpdateProduct(model.Product{ID: 7, Name: "Camiseta"})

	require.ErrorIs(t, err, errRepository)
	assert.Zero(t, updated, "on failure nothing partially filled should leak out")
}

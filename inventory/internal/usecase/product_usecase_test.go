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
	return usecase.NewProductUsecase(repository, &fakeMovementRepository{})
}

func newUsecaseWithMovements(
	repository model.ProductRepository,
	movements model.StockMovementRepository,
) usecase.ProductUsecase {
	return usecase.NewProductUsecase(repository, movements)
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

func TestCreateProductStoresTheInitialStockAsAConfirmedEntry(t *testing.T) {
	repository := &fakeRepository{newID: 42}
	movements := &fakeMovementRepository{newID: 9}
	productUsecase := newUsecaseWithMovements(repository, movements)

	product := model.Product{Code: "PRD-0001", Name: "Camiseta", Unit: "UN", Price: 30.99, Stock: 12}
	created, err := productUsecase.CreateProduct(product)

	require.NoError(t, err)
	assert.Zero(t, repository.receivedProduct.Stock, "the ledger owns the balance, not the insert")
	assert.Zero(t, repository.receivedProduct.ID, "the id is the database's to assign")
	assert.Equal(t, model.StockMovement{
		ProductID: 42,
		Type:      model.MovementIn,
		Origin:    model.MovementOriginAdjustment,
		Quantity:  12,
		Confirmed: true,
	}, movements.receivedMovement)
	assert.Equal(t, 12, created.Stock)
}

func TestCreateProductWithoutInitialStockCreatesNoMovement(t *testing.T) {
	movements := &fakeMovementRepository{}
	productUsecase := newUsecaseWithMovements(&fakeRepository{newID: 42}, movements)

	created, err := productUsecase.CreateProduct(model.Product{Code: "PRD-0001", Name: "Camiseta"})

	require.NoError(t, err)
	assert.Zero(t, movements.calls, "an empty ledger is not an entry of zero")
	assert.Zero(t, created.Stock)
}

func TestCreateProductWhenTheMovementFailsReturnsTheZeroValue(t *testing.T) {
	movements := &fakeMovementRepository{err: errRepository}
	productUsecase := newUsecaseWithMovements(&fakeRepository{newID: 42}, movements)

	created, err := productUsecase.CreateProduct(model.Product{Name: "Camiseta", Stock: 5})

	assert.ErrorIs(t, err, errRepository)
	assert.Equal(t, model.Product{}, created)
}

func TestCreateProductWhenTheRepositoryFailsReturnsTheZeroValue(t *testing.T) {
	productUsecase := newUsecase(&fakeRepository{err: errRepository})

	created, err := productUsecase.CreateProduct(model.Product{Name: "Camiseta"})

	assert.ErrorIs(t, err, errRepository)
	assert.Equal(t, model.Product{}, created, "nothing partially filled leaks out on failure")
}

func TestUpdateProductReturnsTheStoredProductWithItsCurrentStock(t *testing.T) {
	stored := model.Product{ID: 7, Code: "PRD-0007", Name: "Camiseta", Price: 30.99, Stock: 4}
	repository := &fakeRepository{product: stored}
	productUsecase := newUsecase(repository)

	product := model.Product{ID: 7, Code: "PRD-0007", Name: "Camiseta", Price: 30.99}
	updated, err := productUsecase.UpdateProduct(product)

	require.NoError(t, err)
	assert.Equal(t, stored, updated, "the stock comes back from the ledger, not from the request")
	assert.Equal(t, product, repository.receivedProduct)
	assert.Equal(t, 2, repository.calls)
}

func TestUpdateProductPropagatesTheRepositoryError(t *testing.T) {
	repository := &fakeRepository{err: errRepository}
	productUsecase := newUsecase(repository)

	updated, err := productUsecase.UpdateProduct(model.Product{ID: 7, Name: "Camiseta"})

	require.ErrorIs(t, err, errRepository)
	assert.Zero(t, updated, "on failure nothing partially filled should leak out")
}

package usecase

import (
	"inventory/internal/model"
)

type ProductRepository interface {
	GetProducts() ([]model.Product, error)
	GetProductByID(id int) (model.Product, error)
	CreateProduct(product model.Product) (int, error)
}

type ProductUsecase struct {
	repository ProductRepository
}

func NewProductUsecase(repository ProductRepository) ProductUsecase {
	return ProductUsecase{
		repository: repository,
	}
}

func (pu *ProductUsecase) GetProducts() ([]model.Product, error) {
	return pu.repository.GetProducts()
}

func (pu *ProductUsecase) GetProductByID(id int) (model.Product, error) {
	return pu.repository.GetProductByID(id)
}

func (pu *ProductUsecase) CreateProduct(product model.Product) (model.Product, error) {
	productId, err := pu.repository.CreateProduct(product)
	if err != nil {
		return model.Product{}, err
	}

	product.ID = productId

	return product, nil
}

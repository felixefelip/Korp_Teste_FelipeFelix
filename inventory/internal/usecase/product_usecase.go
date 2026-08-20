package usecase

import (
	"inventory/internal/model"
)

type ProductUsecase struct {
	repository model.ProductRepository
}

func NewProductUsecase(repository model.ProductRepository) ProductUsecase {
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

func (pu *ProductUsecase) UpdateProduct(product model.Product) (model.Product, error) {
	if err := pu.repository.UpdateProduct(product); err != nil {
		return model.Product{}, err
	}

	return product, nil
}

func (pu *ProductUsecase) CreateProduct(product model.Product) (model.Product, error) {
	productId, err := pu.repository.CreateProduct(product)
	if err != nil {
		return model.Product{}, err
	}

	product.ID = productId

	return product, nil
}

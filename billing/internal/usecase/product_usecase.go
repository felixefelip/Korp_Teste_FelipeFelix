package usecase

import (
	"billing/internal/model"
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

func (pu *ProductUsecase) SaveProduct(product model.Product) error {
	return pu.repository.UpsertProduct(product)
}

func (pu *ProductUsecase) RemoveProduct(inventoryID int) error {
	return pu.repository.DeactivateProduct(inventoryID)
}

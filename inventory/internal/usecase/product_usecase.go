package usecase

import (
	"inventory/internal/model"
)

type ProductUsecase struct {
	repository         model.ProductRepository
	movementRepository model.StockMovementRepository
}

func NewProductUsecase(
	repository model.ProductRepository,
	movementRepository model.StockMovementRepository,
) ProductUsecase {
	return ProductUsecase{
		repository:         repository,
		movementRepository: movementRepository,
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

	return pu.repository.GetProductByID(product.ID)
}

func (pu *ProductUsecase) CreateProduct(product model.Product) (model.Product, error) {
	initialStock := product.Stock
	product.Stock = 0

	productId, err := pu.repository.CreateProduct(product)
	if err != nil {
		return model.Product{}, err
	}

	product.ID = productId

	if initialStock == 0 {
		return product, nil
	}

	movement := model.StockMovement{
		ProductID: productId,
		Type:      model.MovementIn,
		Origin:    model.MovementOriginAdjustment,
		Quantity:  initialStock,
		Confirmed: true,
	}

	if _, err := pu.movementRepository.CreateMovement(movement); err != nil {
		return model.Product{}, err
	}

	product.Stock = initialStock

	return product, nil
}

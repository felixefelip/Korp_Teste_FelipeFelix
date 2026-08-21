package usecase

import (
	"inventory/internal/model"
)

type StockMovementUsecase struct {
	repository        model.StockMovementRepository
	productRepository model.ProductRepository
}

func NewStockMovementUsecase(
	repository model.StockMovementRepository,
	productRepository model.ProductRepository,
) StockMovementUsecase {
	return StockMovementUsecase{
		repository:        repository,
		productRepository: productRepository,
	}
}

func (su *StockMovementUsecase) GetMovementsByProductID(productID int) ([]model.StockMovement, error) {
	if _, err := su.productRepository.GetProductByID(productID); err != nil {
		return nil, err
	}

	return su.repository.GetMovementsByProductID(productID)
}

func (su *StockMovementUsecase) GetMovementByID(id int) (model.StockMovement, error) {
	return su.repository.GetMovementByID(id)
}

func (su *StockMovementUsecase) CreateMovement(movement model.StockMovement) (model.StockMovement, error) {
	if _, err := su.productRepository.GetProductByID(movement.ProductID); err != nil {
		return model.StockMovement{}, err
	}

	movement.Origin = model.MovementOriginAdjustment
	movement.InvoiceItemID = nil

	movementID, err := su.repository.CreateMovement(movement)
	if err != nil {
		return model.StockMovement{}, err
	}

	movement.ID = movementID

	return movement, nil
}

func (su *StockMovementUsecase) UpdateMovement(movement model.StockMovement) (model.StockMovement, error) {
	stored, err := su.repository.GetMovementByID(movement.ID)
	if err != nil {
		return model.StockMovement{}, err
	}

	if stored.FromInvoice() {
		return model.StockMovement{}, model.ErrMovementFromInvoice
	}

	movement.ProductID = stored.ProductID
	movement.Origin = stored.Origin
	movement.InvoiceItemID = nil

	if err := su.repository.UpdateMovement(movement); err != nil {
		return model.StockMovement{}, err
	}

	return movement, nil
}

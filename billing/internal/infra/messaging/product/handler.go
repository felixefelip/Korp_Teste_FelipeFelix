package product

import (
	"encoding/json"
	"time"

	"billing/internal/model"
	"billing/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

type productEvent struct {
	EventID    string    `json:"eventId"`
	OccurredAt time.Time `json:"occurredAt"`
	ProductID  int       `json:"productId"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Unit       string    `json:"unit"`
	Price      float64   `json:"price"`
}

func (e productEvent) toModel() model.Product {
	return model.Product{
		InventoryID: e.ProductID,
		Code:        e.Code,
		Name:        e.Name,
		Unit:        e.Unit,
		Price:       e.Price,
	}
}

type Handler struct {
	usecase usecase.ProductUsecase
}

func NewHandler(productUsecase usecase.ProductUsecase) *Handler {
	return &Handler{
		usecase: productUsecase,
	}
}

func (h *Handler) HandleProductSaved(delivery amqp.Delivery) error {
	event, err := decode(delivery)
	if err != nil {
		return err
	}

	return h.usecase.SaveProduct(event.toModel())
}

func (h *Handler) HandleProductDeleted(delivery amqp.Delivery) error {
	event, err := decode(delivery)
	if err != nil {
		return err
	}

	return h.usecase.RemoveProduct(event.ProductID)
}

func decode(delivery amqp.Delivery) (productEvent, error) {
	var event productEvent

	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return productEvent{}, err
	}

	return event, nil
}

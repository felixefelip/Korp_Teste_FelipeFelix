package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	OutboxAggregateProduct = "product"

	ProductCreatedKey = "product.created"
	ProductUpdatedKey = "product.updated"
	ProductDeletedKey = "product.deleted"
)

type productPayload struct {
	EventID    string    `json:"eventId"`
	OccurredAt time.Time `json:"occurredAt"`
	ProductID  int       `json:"productId"`
	Code       string    `json:"code,omitempty"`
	Name       string    `json:"name,omitempty"`
	Unit       string    `json:"unit,omitempty"`
	Price      float64   `json:"price,omitempty"`
}

func NewProductCreated(product Product) (OutboxEvent, error) {
	return newProductEvent(ProductCreatedKey, product)
}

func NewProductUpdated(product Product) (OutboxEvent, error) {
	return newProductEvent(ProductUpdatedKey, product)
}

func NewProductDeleted(productID int) (OutboxEvent, error) {
	return newProductEvent(ProductDeletedKey, Product{ID: productID})
}

func newProductEvent(routingKey string, product Product) (OutboxEvent, error) {
	eventID := uuid.NewString()
	occurredAt := time.Now().UTC()

	payload, err := json.Marshal(productPayload{
		EventID:    eventID,
		OccurredAt: occurredAt,
		ProductID:  product.ID,
		Code:       product.Code,
		Name:       product.Name,
		Unit:       product.Unit,
		Price:      product.Price,
	})
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		EventID:       eventID,
		AggregateType: OutboxAggregateProduct,
		AggregateID:   product.ID,
		RoutingKey:    routingKey,
		Payload:       payload,
		CreatedAt:     occurredAt,
		NextAttemptAt: occurredAt,
	}, nil
}

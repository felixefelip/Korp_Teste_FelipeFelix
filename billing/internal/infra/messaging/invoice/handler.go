package invoice

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"billing/internal/model"
	"billing/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

type shortagePayload struct {
	ProductID int    `json:"productId"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Required  int    `json:"required"`
	Available int    `json:"available"`
}

type stockResultEvent struct {
	EventID     string            `json:"eventId"`
	CausationID string            `json:"causationId"`
	OccurredAt  time.Time         `json:"occurredAt"`
	InvoiceID   int               `json:"invoiceId"`
	Reason      string            `json:"reason"`
	Shortages   []shortagePayload `json:"shortages"`
}

type Handler struct {
	usecase usecase.InvoiceUsecase
}

func NewHandler(invoiceUsecase usecase.InvoiceUsecase) *Handler {
	return &Handler{
		usecase: invoiceUsecase,
	}
}

func (h *Handler) HandleStockApplied(delivery amqp.Delivery) error {
	event, err := decode(delivery)
	if err != nil {
		return err
	}

	fmt.Printf("invoice %d: stock applied, caused by %s\n", event.InvoiceID, event.CausationID)

	return settle(h.usecase.ConfirmClose(event.InvoiceID), event.InvoiceID)
}

func (h *Handler) HandleStockRejected(delivery amqp.Delivery) error {
	event, err := decode(delivery)
	if err != nil {
		return err
	}

	fmt.Printf("invoice %d: stock refused (%s), caused by %s\n",
		event.InvoiceID, event.Reason, event.CausationID)

	return settle(h.usecase.RejectClose(event.InvoiceID, event.Reason), event.InvoiceID)
}

func decode(delivery amqp.Delivery) (stockResultEvent, error) {
	var event stockResultEvent

	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return stockResultEvent{}, err
	}

	return event, nil
}

func settle(err error, invoiceID int) error {
	if errors.Is(err, model.ErrInvoiceNotProcessing) {
		fmt.Printf("invoice %d is not being processed, ignoring the result\n", invoiceID)

		return nil
	}

	return err
}

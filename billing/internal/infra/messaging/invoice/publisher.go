package invoice

import (
	"context"
	"encoding/json"
	"time"

	"billing/internal/infra/messaging"
	"billing/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

const publishTimeout = 5 * time.Second

type Publisher struct {
	channel *amqp.Channel
}

func NewPublisher(channel *amqp.Channel) *Publisher {
	return &Publisher{channel: channel}
}

func (p *Publisher) PublishCloseRequested(invoice model.Invoice) error {
	event := newCloseRequestedEvent(invoice)

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()

	return p.channel.PublishWithContext(
		ctx,
		messaging.EventsExchange,
		CloseRequestedKey,
		false,
		false,
		amqp.Publishing{
			MessageId:    event.EventID,
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    event.OccurredAt,
			Body:         body,
		},
	)
}

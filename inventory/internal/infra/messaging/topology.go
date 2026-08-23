package messaging

import (
	"inventory/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

func declareTopology(channel *amqp.Channel) error {
	err := channel.ExchangeDeclare(BillingExchange, amqp.ExchangeTopic, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = channel.ExchangeDeclare(InventoryExchange, amqp.ExchangeTopic, true, false, false, false, nil)
	if err != nil {
		return err
	}

	_, err = channel.QueueDeclare(InvoiceRequestsQueue, true, false, false, false, amqp.Table{
		"x-queue-type": "quorum",
	})
	if err != nil {
		return err
	}

	return bindQueue(channel, InvoiceRequestsQueue, BillingExchange,
		model.InvoiceCloseRequestedKey,
		model.InvoiceReopenRequestedKey,
	)
}

func bindQueue(channel *amqp.Channel, queue, exchange string, routingKeys ...string) error {
	for _, routingKey := range routingKeys {
		if err := channel.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
			return err
		}
	}

	return nil
}

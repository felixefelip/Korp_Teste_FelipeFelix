package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareTopology(channel *amqp.Channel) error {
	err := channel.ExchangeDeclare(BillingExchange, amqp.ExchangeTopic, true, false, false, false, nil)
	if err != nil {
		return err
	}

	_, err = channel.QueueDeclare(InvoiceRequestsQueue, true, false, false, false, amqp.Table{
		"x-queue-type": "quorum",
	})
	if err != nil {
		return err
	}

	return channel.QueueBind(InvoiceRequestsQueue, InvoiceRequestsKey, BillingExchange, false, nil)
}

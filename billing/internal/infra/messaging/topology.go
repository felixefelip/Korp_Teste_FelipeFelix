package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareTopology(channel *amqp.Channel) error {
	return channel.ExchangeDeclare(EventsExchange, amqp.ExchangeTopic, true, false, false, false, nil)
}

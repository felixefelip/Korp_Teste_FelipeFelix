package messaging

import (
	"billing/internal/model"

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

	if err := declareDeadLetter(channel); err != nil {
		return err
	}

	_, err = channel.QueueDeclare(StockResultsQueue, true, false, false, false, consumedQueueArguments())
	if err != nil {
		return err
	}

	_, err = channel.QueueDeclare(CatalogQueue, true, false, false, false, consumedQueueArguments())
	if err != nil {
		return err
	}

	err = bindQueue(channel, CatalogQueue, InventoryExchange,
		model.ProductCreatedKey,
		model.ProductUpdatedKey,
		model.ProductDeletedKey,
	)
	if err != nil {
		return err
	}

	return bindQueue(channel, StockResultsQueue, InventoryExchange,
		model.InvoiceStockAppliedKey,
		model.InvoiceStockRejectedKey,
		model.InvoiceStockRevertedKey,
		model.InvoiceStockRevertRejectedKey,
	)
}

func consumedQueueArguments() amqp.Table {
	return amqp.Table{
		"x-queue-type":           "quorum",
		"x-dead-letter-exchange": DeadLetterExchange,
	}
}

func declareDeadLetter(channel *amqp.Channel) error {
	err := channel.ExchangeDeclare(DeadLetterExchange, amqp.ExchangeTopic, true, false, false, false, nil)
	if err != nil {
		return err
	}

	_, err = channel.QueueDeclare(DeadLetterQueue, true, false, false, false, amqp.Table{
		"x-queue-type": "quorum",
	})
	if err != nil {
		return err
	}

	return bindQueue(channel, DeadLetterQueue, DeadLetterExchange, "#")
}

func bindQueue(channel *amqp.Channel, queue, exchange string, routingKeys ...string) error {
	for _, routingKey := range routingKeys {
		if err := channel.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
			return err
		}
	}

	return nil
}

package messaging

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	BillingExchange      = "billing.events"
	InventoryExchange    = "inventory.events"
	InvoiceRequestsQueue = "inventory.invoice-requests"

	DeadLetterExchange = "inventory.dead-letter"
	DeadLetterQueue    = "inventory.dead-letter"

	prefetchCount = 1
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func URL() string {
	return env("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
}

type Connection struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	returns    chan amqp.Return
}

func Connect() (*Connection, error) {
	connection, err := amqp.Dial(URL())
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()

		return nil, err
	}

	if err := declareTopology(channel); err != nil {
		connection.Close()

		return nil, err
	}

	if err := channel.Qos(prefetchCount, 0, false); err != nil {
		connection.Close()

		return nil, err
	}

	if err := channel.Confirm(false); err != nil {
		connection.Close()

		return nil, err
	}

	returns := channel.NotifyReturn(make(chan amqp.Return, 1))

	return &Connection{connection: connection, channel: channel, returns: returns}, nil
}

func (c *Connection) Closed() bool {
	return c.connection.IsClosed()
}

func (c *Connection) Close() error {
	return c.connection.Close()
}

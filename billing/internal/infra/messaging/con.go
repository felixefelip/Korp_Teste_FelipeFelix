package messaging

import (
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange = "billing.events"

	connectAttempts = 10
	connectInterval = 3 * time.Second
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
}

func Connect() (*Connection, error) {
	var lastErr error

	for attempt := 1; attempt <= connectAttempts; attempt++ {
		connection, err := amqp.Dial(URL())
		if err != nil {
			lastErr = err

			fmt.Printf("RabbitMQ unavailable (%d/%d): %v\n", attempt, connectAttempts, err)
			time.Sleep(connectInterval)

			continue
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

		fmt.Println("Successfully connected to RabbitMQ")

		return &Connection{connection: connection, channel: channel}, nil
	}

	return nil, lastErr
}

func (c *Connection) Channel() *amqp.Channel {
	return c.channel
}

func (c *Connection) Close() error {
	return c.connection.Close()
}

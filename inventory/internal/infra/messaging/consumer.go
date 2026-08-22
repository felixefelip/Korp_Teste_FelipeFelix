package messaging

import (
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const reconnectInterval = 3 * time.Second

var errDeliveriesClosed = errors.New("connection to the broker was lost")

type Handler func(delivery amqp.Delivery) error

type Consumer struct {
	queue  string
	handle Handler
}

func NewConsumer(queue string, handle Handler) *Consumer {
	return &Consumer{
		queue:  queue,
		handle: handle,
	}
}

func (c *Consumer) Start() {
	go func() {
		for {
			if err := c.consume(); err != nil {
				fmt.Printf("consumer of %s: %v\n", c.queue, err)
			}

			time.Sleep(reconnectInterval)
		}
	}()

	fmt.Println("Consumer of " + c.queue + " started")
}

func (c *Consumer) consume() error {
	connection, err := Connect()
	if err != nil {
		return err
	}
	defer connection.Close()

	deliveries, err := connection.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	fmt.Println("Consuming " + c.queue)

	for delivery := range deliveries {
		if err := c.handle(delivery); err != nil {
			fmt.Printf("message %s refused: %v\n", delivery.MessageId, err)
			delivery.Nack(false, false)

			continue
		}

		delivery.Ack(false)
	}

	return errDeliveriesClosed
}

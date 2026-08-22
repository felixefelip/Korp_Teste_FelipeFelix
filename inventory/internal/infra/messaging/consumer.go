package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler func(delivery amqp.Delivery) error

func (c *Connection) Consume(queue string, handle Handler) error {
	deliveries, err := c.channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveries {
			if err := handle(delivery); err != nil {
				fmt.Printf("message %s refused: %v\n", delivery.MessageId, err)
				delivery.Nack(false, false)

				continue
			}

			delivery.Ack(false)
		}
	}()

	fmt.Println("Consuming " + queue)

	return nil
}

package messaging

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const reconnectInterval = 3 * time.Second

var errDeliveriesClosed = errors.New("connection to the broker was lost")

type Handler func(delivery amqp.Delivery) error

type Routes map[string]Handler

func (r Routes) keys() []string {
	keys := make([]string, 0, len(r))

	for key := range r {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

type Consumer struct {
	queue  string
	routes Routes
}

func NewConsumer(queue string, routes Routes) *Consumer {
	return &Consumer{
		queue:  queue,
		routes: routes,
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

	fmt.Printf("Consumer of %s started, handling %s\n", c.queue, strings.Join(c.routes.keys(), ", "))
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
		c.dispatch(delivery)
	}

	return errDeliveriesClosed
}

func (c *Consumer) dispatch(delivery amqp.Delivery) {
	handle, known := c.routes[delivery.RoutingKey]
	if !known {
		fmt.Printf("no handler for %s, discarding message %s\n", delivery.RoutingKey, delivery.MessageId)
		delivery.Ack(false)

		return
	}

	if err := handle(delivery); err != nil {
		fmt.Printf("message %s refused: %v\n", delivery.MessageId, err)
		delivery.Nack(false, false)

		return
	}

	delivery.Ack(false)
}

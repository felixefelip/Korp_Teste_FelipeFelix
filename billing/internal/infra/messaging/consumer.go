package messaging

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"billing/internal/infra/messaging/msgerr"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	reconnectInterval = 3 * time.Second

	retryLimit       = 10
	retryBackoffBase = 2 * time.Second
	retryBackoffCap  = 30 * time.Second
)

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
	wait   func(time.Duration)
}

func NewConsumer(queue string, routes Routes) *Consumer {
	return &Consumer{
		queue:  queue,
		routes: routes,
		wait:   time.Sleep,
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

	err := handle(delivery)
	if err == nil {
		delivery.Ack(false)

		return
	}

	if msgerr.IsPoison(err) {
		fmt.Printf("message %s dead-lettered without retrying: %v\n", delivery.MessageId, err)
		delivery.Nack(false, false)

		return
	}

	attempts := acquiredCount(delivery)

	if attempts >= retryLimit {
		fmt.Printf("message %s dead-lettered after %d attempts: %v\n", delivery.MessageId, attempts, err)
		delivery.Nack(false, false)

		return
	}

	delay := retryDelay(attempts)

	fmt.Printf("message %s returned to %s, retrying in %s: %v\n", delivery.MessageId, c.queue, delay, err)
	c.wait(delay)
	delivery.Nack(false, true)
}

func retryDelay(attempts int) time.Duration {
	delay := retryBackoffBase << attempts

	if delay > retryBackoffCap || delay <= 0 {
		return retryBackoffCap
	}

	return delay
}

func acquiredCount(delivery amqp.Delivery) int {
	for _, header := range []string{"x-acquired-count", "x-delivery-count"} {
		switch count := delivery.Headers[header].(type) {
		case int32:
			return int(count)
		case int64:
			return int(count)
		case int:
			return count
		}
	}

	return 0
}

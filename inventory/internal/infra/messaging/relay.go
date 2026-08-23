package messaging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"inventory/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	relayInterval  = time.Second
	relayBatch     = 50
	relayLease     = 30 * time.Second
	publishTimeout = 5 * time.Second

	backoffBase = 2 * time.Second
	backoffCap  = time.Minute
)

var (
	errBrokerRefused = errors.New("broker refused the message")
	errUnroutable    = errors.New("no queue bound to the routing key")
)

type Relay struct {
	repository model.OutboxRepository
	exchange   string
	connection *Connection
}

func NewRelay(repository model.OutboxRepository, exchange string) *Relay {
	return &Relay{
		repository: repository,
		exchange:   exchange,
	}
}

func (r *Relay) Start() {
	go func() {
		ticker := time.NewTicker(relayInterval)
		defer ticker.Stop()

		for range ticker.C {
			r.drain()
		}
	}()

	fmt.Println("Outbox relay started, publishing to " + r.exchange)
}

func (r *Relay) drain() {
	if !r.connected() {
		return
	}

	events, err := r.repository.ClaimEvents(relayBatch, relayLease)
	if err != nil {
		fmt.Println("claiming outbox events:", err)

		return
	}

	for _, event := range events {
		if err := r.publish(event); err != nil {
			fmt.Printf("publishing event %s: %v\n", event.EventID, err)

			r.fail(event, err)
			r.disconnect()

			return
		}

		if err := r.repository.MarkPublished(event.ID); err != nil {
			fmt.Println("marking event as published:", err)
		}
	}
}

func (r *Relay) connected() bool {
	if r.connection != nil && !r.connection.Closed() {
		return true
	}

	connection, err := Connect()
	if err != nil {
		return false
	}

	fmt.Println("Successfully connected to RabbitMQ")

	r.connection = connection

	return true
}

func (r *Relay) disconnect() {
	if r.connection == nil {
		return
	}

	r.connection.Close()
	r.connection = nil
}

func (r *Relay) publish(event model.OutboxEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()

	confirmation, err := r.connection.channel.PublishWithDeferredConfirmWithContext(
		ctx,
		r.exchange,
		event.RoutingKey,
		true,
		false,
		amqp.Publishing{
			MessageId:     event.EventID,
			CorrelationId: event.CausationID,
			ContentType:   "application/json",
			DeliveryMode:  amqp.Persistent,
			Timestamp:     event.CreatedAt,
			Body:          event.Payload,
		},
	)
	if err != nil {
		return err
	}

	acknowledged, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}

	if !acknowledged {
		return errBrokerRefused
	}

	select {
	case returned := <-r.connection.returns:
		return fmt.Errorf("%w: %s", errUnroutable, returned.RoutingKey)
	default:
	}

	return nil
}

func (r *Relay) fail(event model.OutboxEvent, cause error) {
	if err := r.repository.MarkFailed(event.ID, cause.Error(), time.Now().Add(backoff(event.Attempts))); err != nil {
		fmt.Println("marking event as failed:", err)
	}
}

func backoff(attempts int) time.Duration {
	delay := backoffBase << attempts

	if delay > backoffCap || delay <= 0 {
		return backoffCap
	}

	return delay
}

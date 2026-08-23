package messaging

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/stretchr/testify/assert"
)

type fakeAcknowledger struct {
	acked   bool
	nacked  bool
	requeue bool
}

func (f *fakeAcknowledger) Ack(tag uint64, multiple bool) error {
	f.acked = true
	return nil
}

func (f *fakeAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	f.nacked = true
	f.requeue = requeue
	return nil
}

func (f *fakeAcknowledger) Reject(tag uint64, requeue bool) error {
	return nil
}

func deliveryOf(routingKey string) (amqp.Delivery, *fakeAcknowledger) {
	acknowledger := &fakeAcknowledger{}

	return amqp.Delivery{
		Acknowledger: acknowledger,
		RoutingKey:   routingKey,
		MessageId:    "message-1",
	}, acknowledger
}

func TestDispatchRunsTheHandlerOfTheRoutingKey(t *testing.T) {
	handled := false
	consumer := NewConsumer("queue", Routes{
		"invoice.close.requested": func(amqp.Delivery) error {
			handled = true
			return nil
		},
	})

	delivery, acknowledger := deliveryOf("invoice.close.requested")
	consumer.dispatch(delivery)

	assert.True(t, handled)
	assert.True(t, acknowledger.acked)
	assert.False(t, acknowledger.nacked)
}

func TestDispatchDiscardsAnUnknownRoutingKey(t *testing.T) {
	handled := false
	consumer := NewConsumer("queue", Routes{
		"invoice.close.requested": func(amqp.Delivery) error {
			handled = true
			return nil
		},
	})

	delivery, acknowledger := deliveryOf("invoice.something.else")
	consumer.dispatch(delivery)

	assert.False(t, handled)
	assert.True(t, acknowledger.acked, "an unknown key is acknowledged, not retried forever")
	assert.False(t, acknowledger.nacked)
}

func TestDispatchNacksWhenTheHandlerFails(t *testing.T) {
	consumer := NewConsumer("queue", Routes{
		"invoice.close.requested": func(amqp.Delivery) error {
			return errors.New("database down")
		},
	})

	delivery, acknowledger := deliveryOf("invoice.close.requested")
	consumer.dispatch(delivery)

	assert.True(t, acknowledger.nacked)
	assert.False(t, acknowledger.requeue, "requeue is the broker's job through the delivery limit")
	assert.False(t, acknowledger.acked)
}

func TestRoutesKeysComeSorted(t *testing.T) {
	routes := Routes{
		"invoice.reopen.requested": nil,
		"invoice.close.requested":  nil,
	}

	assert.Equal(t, []string{"invoice.close.requested", "invoice.reopen.requested"}, routes.keys())
}

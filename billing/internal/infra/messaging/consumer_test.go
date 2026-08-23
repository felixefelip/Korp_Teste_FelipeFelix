package messaging

import (
	"errors"
	"testing"
	"time"

	"billing/internal/infra/messaging/msgerr"

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

func consumerOf(routes Routes) (*Consumer, *time.Duration) {
	consumer := NewConsumer("queue", routes)
	waited := time.Duration(0)

	consumer.wait = func(delay time.Duration) {
		waited = delay
	}

	return consumer, &waited
}

func TestDispatchRunsTheHandlerOfTheRoutingKey(t *testing.T) {
	handled := false
	consumer, _ := consumerOf(Routes{
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
	consumer, _ := consumerOf(Routes{
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

func TestDispatchReturnsTheMessageWhenTheFailureIsTransient(t *testing.T) {
	consumer, waited := consumerOf(Routes{
		"invoice.close.requested": func(amqp.Delivery) error {
			return errors.New("database down")
		},
	})

	delivery, acknowledger := deliveryOf("invoice.close.requested")
	consumer.dispatch(delivery)

	assert.True(t, acknowledger.nacked)
	assert.True(t, acknowledger.requeue, "a database that came back makes the next delivery work")
	assert.False(t, acknowledger.acked)
	assert.Equal(t, retryBackoffBase, *waited, "the queue is not hammered between deliveries")
}

func TestDispatchDiscardsAPoisonMessage(t *testing.T) {
	consumer, waited := consumerOf(Routes{
		"invoice.close.requested": func(amqp.Delivery) error {
			return msgerr.Poison(errors.New("invalid character 'n'"))
		},
	})

	delivery, acknowledger := deliveryOf("invoice.close.requested")
	consumer.dispatch(delivery)

	assert.True(t, acknowledger.nacked)
	assert.False(t, acknowledger.requeue, "no delivery will ever decode it")
	assert.Zero(t, *waited, "there is nothing to wait for")
}

func TestDispatchGivesUpOnAMessageThatExhaustedTheRetries(t *testing.T) {
	consumer, waited := consumerOf(Routes{
		"invoice.close.requested": func(amqp.Delivery) error {
			return errors.New("database down")
		},
	})

	delivery, acknowledger := deliveryOf("invoice.close.requested")
	delivery.Headers = amqp.Table{"x-acquired-count": int64(retryLimit)}
	consumer.dispatch(delivery)

	assert.True(t, acknowledger.nacked)
	assert.False(t, acknowledger.requeue, "the broker does not bound explicit returns, the consumer does")
	assert.Zero(t, *waited, "it is being dead-lettered, not retried")
}

func TestAcquiredCountReadsWhatTheBrokerCounted(t *testing.T) {
	delivery, _ := deliveryOf("invoice.close.requested")

	assert.Zero(t, acquiredCount(delivery), "the first delivery carries no header")

	delivery.Headers = amqp.Table{"x-acquired-count": int64(3)}
	assert.Equal(t, 3, acquiredCount(delivery))

	delivery.Headers = amqp.Table{"x-delivery-count": int32(4)}
	assert.Equal(t, 4, acquiredCount(delivery), "older brokers count under the other name")
}

func TestRetryDelayGrowsWithTheAttemptsAndStopsAtTheCap(t *testing.T) {
	assert.Equal(t, retryBackoffBase, retryDelay(0))
	assert.Equal(t, 4*retryBackoffBase, retryDelay(2))
	assert.Equal(t, retryBackoffCap, retryDelay(retryLimit))
}

func TestRoutesKeysComeSorted(t *testing.T) {
	routes := Routes{
		"invoice.reopen.requested": nil,
		"invoice.close.requested":  nil,
	}

	assert.Equal(t, []string{"invoice.close.requested", "invoice.reopen.requested"}, routes.keys())
}

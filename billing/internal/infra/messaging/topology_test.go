package messaging

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	probeExchange  = "billing.test.dead-letter"
	probeQueue     = "billing.test.retries"
	probeDeadQueue = "billing.test.dead-letters"
)

func TestConsumedQueuesAskTheBrokerForTheDeadLetterExchange(t *testing.T) {
	arguments := consumedQueueArguments()

	assert.Equal(t, "quorum", arguments["x-queue-type"])
	assert.Equal(t, DeadLetterExchange, arguments["x-dead-letter-exchange"])
}

func newProbeChannel(t *testing.T) *amqp.Channel {
	t.Helper()

	connection, err := amqp.Dial(URL())
	require.NoError(t, err)

	channel, err := connection.Channel()
	require.NoError(t, err)

	t.Cleanup(func() {
		channel.QueueDelete(probeQueue, false, false, false)
		channel.QueueDelete(probeDeadQueue, false, false, false)
		channel.ExchangeDelete(probeExchange, false, false)
		connection.Close()
	})

	require.NoError(t, channel.ExchangeDeclare(probeExchange, amqp.ExchangeTopic, true, false, false, false, nil))

	_, err = channel.QueueDeclare(probeDeadQueue, true, false, false, false, amqp.Table{
		"x-queue-type": "quorum",
	})
	require.NoError(t, err)

	require.NoError(t, channel.QueueBind(probeDeadQueue, "#", probeExchange, false, nil))

	_, err = channel.QueueDeclare(probeQueue, true, false, false, false, amqp.Table{
		"x-queue-type":           "quorum",
		"x-dead-letter-exchange": probeExchange,
	})
	require.NoError(t, err)

	require.NoError(t, channel.Qos(prefetchCount, 0, false))

	return channel
}

func publishProbe(t *testing.T, channel *amqp.Channel) {
	t.Helper()

	require.NoError(t, channel.Publish("", probeQueue, false, false, amqp.Publishing{
		MessageId:    "probe-1",
		DeliveryMode: amqp.Persistent,
		Body:         []byte(`{"invoiceId":1}`),
	}))
}

func nextProbe(t *testing.T, deliveries <-chan amqp.Delivery) amqp.Delivery {
	t.Helper()

	select {
	case delivery := <-deliveries:
		return delivery
	case <-time.After(5 * time.Second):
		t.Fatal("the broker did not deliver the message")

		return amqp.Delivery{}
	}
}

func deadLettered(t *testing.T, channel *amqp.Channel) (amqp.Delivery, bool) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		delivery, found, err := channel.Get(probeDeadQueue, true)
		require.NoError(t, err)

		if found {
			return delivery, true
		}

		time.Sleep(100 * time.Millisecond)
	}

	return amqp.Delivery{}, false
}

func TestAReturnedMessageComesBackAndTheBrokerCountsIt(t *testing.T) {
	channel := newProbeChannel(t)

	deliveries, err := channel.Consume(probeQueue, "", false, false, false, false, nil)
	require.NoError(t, err)

	publishProbe(t, channel)

	counts := make([]int, 0, 3)

	for range 3 {
		delivery := nextProbe(t, deliveries)

		counts = append(counts, acquiredCount(delivery))
		require.NoError(t, delivery.Nack(false, true))
	}

	assert.Equal(t, []int{0, 1, 2}, counts, "the count the consumer budgets its retries against")

	_, found, err := channel.Get(probeDeadQueue, true)
	require.NoError(t, err)
	assert.False(t, found, "the broker never dead-letters an explicit return, however many there are")
}

func TestARefusedMessageIsDeadLetteredWithoutBeingDeliveredAgain(t *testing.T) {
	channel := newProbeChannel(t)

	deliveries, err := channel.Consume(probeQueue, "", false, false, false, false, nil)
	require.NoError(t, err)

	publishProbe(t, channel)

	require.NoError(t, nextProbe(t, deliveries).Nack(false, false))

	dead, found := deadLettered(t, channel)

	require.True(t, found, "a refused message is kept, not discarded")
	assert.Equal(t, "probe-1", dead.MessageId)

	select {
	case delivery := <-deliveries:
		t.Fatalf("the message was delivered again as %s", delivery.MessageId)
	case <-time.After(time.Second):
	}
}

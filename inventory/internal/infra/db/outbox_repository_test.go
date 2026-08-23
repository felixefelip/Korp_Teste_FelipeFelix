package db_test

import (
	"testing"
	"time"

	"inventory/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testLease = time.Minute

func TestClaimEventsReturnsWhatIsStillPending(t *testing.T) {
	repository := newOutboxRepository(t)
	store(t, pendingEvent(1), pendingEvent(2))

	events, err := repository.ClaimEvents(10, testLease)
	require.NoError(t, err)

	require.Len(t, events, 2)
	assert.Equal(t, 1, events[0].AggregateID)
	assert.Equal(t, 2, events[1].AggregateID, "oldest first")
}

func TestClaimEventsLeasesWhatItHandsOut(t *testing.T) {
	repository := newOutboxRepository(t)
	store(t, pendingEvent(1))

	_, err := repository.ClaimEvents(10, testLease)
	require.NoError(t, err)

	again, err := repository.ClaimEvents(10, testLease)
	require.NoError(t, err)

	assert.Empty(t, again, "a second relay does not pick up what is already in flight")
}

func TestClaimEventsHonoursTheLimit(t *testing.T) {
	repository := newOutboxRepository(t)
	store(t, pendingEvent(1), pendingEvent(2), pendingEvent(3))

	events, err := repository.ClaimEvents(2, testLease)
	require.NoError(t, err)

	assert.Len(t, events, 2)
}

func TestMarkPublishedTakesTheEventOutOfTheQueue(t *testing.T) {
	repository := newOutboxRepository(t)
	store(t, pendingEvent(1))

	events, err := repository.ClaimEvents(10, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NoError(t, repository.MarkPublished(events[0].ID))

	var stored model.OutboxEvent
	require.NoError(t, testConnection.First(&stored, events[0].ID).Error)
	assert.True(t, stored.Published())

	again, err := repository.ClaimEvents(10, testLease)
	require.NoError(t, err)
	assert.Empty(t, again)
}

func TestMarkFailedCountsTheAttemptAndDelaysTheNextOne(t *testing.T) {
	repository := newOutboxRepository(t)
	store(t, pendingEvent(1))

	events, err := repository.ClaimEvents(10, 0)
	require.NoError(t, err)

	nextAttempt := time.Now().Add(time.Hour)
	require.NoError(t, repository.MarkFailed(events[0].ID, "broker down", nextAttempt))

	var stored model.OutboxEvent
	require.NoError(t, testConnection.First(&stored, events[0].ID).Error)

	assert.Equal(t, 1, stored.Attempts)
	assert.Equal(t, "broker down", stored.LastError)
	assert.False(t, stored.Published())

	again, err := repository.ClaimEvents(10, testLease)
	require.NoError(t, err)
	assert.Empty(t, again, "it waits for the backoff before being retried")
}

func TestMarkPublishedWhenMissingReturnsAnError(t *testing.T) {
	repository := newOutboxRepository(t)

	assert.Error(t, repository.MarkPublished(9999))
}

func pendingEvent(aggregateID int) model.OutboxEvent {
	event, err := model.NewInvoiceStockApplied(aggregateID)
	if err != nil {
		panic(err)
	}

	return event
}

func store(t *testing.T, events ...model.OutboxEvent) {
	t.Helper()

	for _, event := range events {
		require.NoError(t, testConnection.Create(&event).Error)
	}
}

<?php

use App\Messaging\Relay;
use App\Messaging\Topology;
use App\Models\OutboxEvent;
use Illuminate\Support\Carbon;

beforeEach(function () {
    $this->routingKey = 'finance.entry.created';
    $this->scratch = declareScratchTopology($this->routingKey);

    broker()->channel()->queue_bind(
        $this->scratch['queue'], Topology::FINANCE_EXCHANGE, $this->routingKey
    );
});

afterEach(function () {
    deleteScratchTopology($this->scratch['queue']);
});

function relay(): Relay
{
    return Relay::make(log: fn (string $line) => null, sleep: fn (int $seconds) => null);
}

it('publishes a pending event and marks it published', function () {
    $event = OutboxEvent::record(OutboxEvent::AGGREGATE_INVOICE, 42, $this->routingKey, ['invoiceId' => 42]);

    relay()->drain();

    $message = popMessage($this->scratch['queue']);

    expect($message)->not->toBeNull()
        ->and(json_decode($message->getBody(), true)['invoiceId'])->toBe(42)
        ->and($message->get('message_id'))->toBe($event->event_id)
        ->and($message->get('delivery_mode'))->toBe(2)
        ->and($event->fresh()->published())->toBeTrue()
        ->and($event->fresh()->last_error)->toBeNull();
});

it('carries the causation id as the amqp correlation id', function () {
    $causationId = (string) Str::uuid();

    OutboxEvent::record(OutboxEvent::AGGREGATE_INVOICE, 42, $this->routingKey, [], $causationId);

    relay()->drain();

    expect(popMessage($this->scratch['queue'])->get('correlation_id'))->toBe($causationId);
});

it('does not publish the same event twice', function () {
    OutboxEvent::record(OutboxEvent::AGGREGATE_INVOICE, 42, $this->routingKey, []);

    relay()->drain();
    popMessage($this->scratch['queue']);

    relay()->drain();

    expect(popMessage($this->scratch['queue'], timeout: 1.0))->toBeNull();
});

it('marks an event with no queue bound to its key as failed', function () {
    $event = OutboxEvent::record(OutboxEvent::AGGREGATE_INVOICE, 42, 'finance.nobody.listens', []);

    relay()->drain();

    $event = $event->fresh();

    expect($event->published())->toBeFalse()
        ->and($event->attempts)->toBe(1)
        ->and($event->last_error)->toContain('no queue bound to the routing key')
        ->and($event->next_attempt_at->greaterThan(Carbon::now()))->toBeTrue();
});

it('does not retry a failed event before its backoff elapses', function () {
    $event = OutboxEvent::record(OutboxEvent::AGGREGATE_INVOICE, 42, 'finance.nobody.listens', []);

    relay()->drain();
    $event->refresh()->forceFill(['routing_key' => $this->routingKey])->save();

    relay()->drain();

    expect($event->fresh()->published())->toBeFalse()
        ->and($event->fresh()->attempts)->toBe(1);
});

it('retries a failed event once its backoff elapses', function () {
    $event = OutboxEvent::record(OutboxEvent::AGGREGATE_INVOICE, 42, 'finance.nobody.listens', []);

    relay()->drain();
    $event->refresh()->forceFill(['routing_key' => $this->routingKey])->save();

    $this->travel(3)->seconds();
    relay()->drain();

    expect($event->fresh()->published())->toBeTrue()
        ->and(popMessage($this->scratch['queue']))->not->toBeNull();
});

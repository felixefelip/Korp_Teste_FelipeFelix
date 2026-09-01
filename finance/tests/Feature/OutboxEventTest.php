<?php

use App\Models\OutboxEvent;
use Illuminate\Support\Carbon;
use Illuminate\Support\Str;

function recordEvent(string $routingKey = 'finance.entry.created'): OutboxEvent
{
    return OutboxEvent::record(OutboxEvent::AGGREGATE_INVOICE, 42, $routingKey, ['invoiceId' => 42]);
}

it('wraps the payload in the shared envelope', function () {
    $event = recordEvent();

    expect($event->payload)->toHaveKeys(['eventId', 'occurredAt', 'invoiceId'])
        ->and($event->payload['eventId'])->toBe($event->event_id)
        ->and($event->payload['invoiceId'])->toBe(42)
        ->and($event->published())->toBeFalse();
});

it('carries the causation id into the payload and the column', function () {
    $causationId = (string) Str::uuid();

    $event = OutboxEvent::record(OutboxEvent::AGGREGATE_INVOICE, 1, 'finance.entry.created', [], $causationId);

    expect($event->causation_id)->toBe($causationId)
        ->and($event->payload['causationId'])->toBe($causationId);
});

it('leases claimed events so a second relay skips them', function () {
    recordEvent();

    expect(OutboxEvent::claim(50, 30))->toHaveCount(1)
        ->and(OutboxEvent::claim(50, 30))->toHaveCount(0);
});

it('claims again once the lease expires', function () {
    $event = recordEvent();

    OutboxEvent::claim(50, 30);
    $event->forceFill(['next_attempt_at' => Carbon::now()->subSecond()])->save();

    expect(OutboxEvent::claim(50, 30))->toHaveCount(1);
});

it('never claims a published event', function () {
    recordEvent()->markPublished();

    expect(OutboxEvent::claim(50, 30))->toHaveCount(0);
});

it('records the failure cause and postpones the next attempt', function () {
    $event = recordEvent();
    $nextAttemptAt = Carbon::now()->addMinute();

    $event->markFailed('broker is down', $nextAttemptAt);

    expect($event->fresh()->attempts)->toBe(1)
        ->and($event->fresh()->last_error)->toBe('broker is down')
        ->and($event->fresh()->next_attempt_at->timestamp)->toBe($nextAttemptAt->timestamp);
});

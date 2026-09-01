<?php

namespace App\Messaging;

use App\Models\OutboxEvent;
use Closure;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Log;
use PhpAmqpLib\Message\AMQPMessage;
use Throwable;

class Relay
{
    private ?Connection $connection = null;

    public function __construct(
        private readonly Closure $log,
        private readonly Closure $sleep,
    ) {}

    public static function make(?Closure $log = null, ?Closure $sleep = null): self
    {
        return new self(
            $log ?? fn (string $line) => Log::info($line),
            $sleep ?? fn (int $seconds) => sleep($seconds),
        );
    }

    public function run(): void
    {
        $interval = (int) config('rabbitmq.relay.interval_seconds');

        $signals = Signals::trap();

        $this->report('outbox relay started, publishing to '.Topology::FINANCE_EXCHANGE);

        while (! $signals->stopping()) {
            $this->drain();

            if ($signals->stopping()) {
                break;
            }

            ($this->sleep)($interval);
        }

        $this->disconnect();
        $this->report('outbox relay stopped');
    }

    public function drain(): void
    {
        if (! $this->connected()) {
            return;
        }

        $events = OutboxEvent::claim(
            (int) config('rabbitmq.relay.batch'),
            (int) config('rabbitmq.relay.lease_seconds'),
        );

        foreach ($events as $event) {
            try {
                $this->publish($event);
            } catch (Throwable $exception) {
                $this->report("publishing event {$event->event_id}: {$exception->getMessage()}");
                $this->fail($event, $exception);
                $this->disconnect();

                return;
            }

            $event->markPublished();
        }
    }

    private function connected(): bool
    {
        if ($this->connection !== null && $this->connection->isOpen()) {
            return true;
        }

        try {
            $this->connection = Connection::open();
        } catch (Throwable $exception) {
            $this->report('connecting to RabbitMQ: '.$exception->getMessage());

            return false;
        }

        $this->report('successfully connected to RabbitMQ');

        return true;
    }

    private function disconnect(): void
    {
        $this->connection?->close();
        $this->connection = null;
    }

    private function publish(OutboxEvent $event): void
    {
        $properties = array_filter([
            'message_id' => $event->event_id,
            'correlation_id' => $event->causation_id,
            'content_type' => 'application/json',
            'delivery_mode' => AMQPMessage::DELIVERY_MODE_PERSISTENT,
            'timestamp' => $event->created_at->getTimestamp(),
        ], fn (mixed $value): bool => $value !== null);

        $this->connection?->publish(
            new AMQPMessage((string) json_encode($event->payload), $properties),
            Topology::FINANCE_EXCHANGE,
            $event->routing_key,
            (float) config('rabbitmq.relay.publish_timeout_seconds'),
        );
    }

    private function fail(OutboxEvent $event, Throwable $cause): void
    {
        $event->markFailed($cause->getMessage(), Carbon::now()->addSeconds($this->backoff($event->attempts)));
    }

    private function backoff(int $attempts): int
    {
        $base = (int) config('rabbitmq.relay.backoff_base_seconds');
        $cap = (int) config('rabbitmq.relay.backoff_cap_seconds');

        $delay = $base * (2 ** $attempts);

        return $delay > $cap || $delay <= 0 ? $cap : (int) $delay;
    }

    private function report(string $line): void
    {
        ($this->log)($line);
    }
}

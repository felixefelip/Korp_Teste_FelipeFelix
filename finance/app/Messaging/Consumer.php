<?php

namespace App\Messaging;

use Closure;
use Illuminate\Support\Facades\Log;
use PhpAmqpLib\Message\AMQPMessage;
use Throwable;

class Consumer
{
    /**
     * @param  array<string, callable(AMQPMessage): void>  $routes
     */
    public function __construct(
        private readonly string $queue,
        private readonly array $routes,
        private readonly Closure $log,
        private readonly Closure $sleep,
    ) {}

    /**
     * @param  array<string, callable(AMQPMessage): void>  $routes
     */
    public static function for(string $queue, array $routes, ?Closure $log = null, ?Closure $sleep = null): self
    {
        return new self(
            $queue,
            $routes,
            $log ?? fn (string $line) => Log::info($line),
            $sleep ?? fn (int $seconds) => sleep($seconds),
        );
    }

    public function run(): void
    {
        $reconnect = (int) config('rabbitmq.consumer.reconnect_seconds');
        $signals = Signals::trap();

        while (! $signals->stopping()) {
            try {
                $this->consume(stopWhen: fn (): bool => $signals->stopping());
            } catch (Throwable $exception) {
                $this->report("consumer of {$this->queue}: {$exception->getMessage()}");
            }

            if ($signals->stopping()) {
                break;
            }

            ($this->sleep)($reconnect);
        }

        $this->report("consumer of {$this->queue} stopped");
    }

    public function consume(?int $stopAfter = null, float $waitTimeout = 0, ?Closure $stopWhen = null): void
    {
        $connection = Connection::open();

        try {
            $handled = 0;
            $channel = $connection->channel();

            $channel->basic_consume(
                $this->queue, '', false, false, false, false,
                function (AMQPMessage $message) use (&$handled): void {
                    $this->dispatch($message);
                    $handled++;
                }
            );

            $this->report("consuming {$this->queue}");

            while ($channel->is_consuming()) {
                $channel->wait(null, false, $waitTimeout);

                if ($stopAfter !== null && $handled >= $stopAfter) {
                    return;
                }

                if ($stopWhen !== null && $stopWhen()) {
                    return;
                }
            }
        } finally {
            $connection->close();
        }
    }

    public function dispatch(AMQPMessage $message): void
    {
        $routingKey = $message->getRoutingKey();
        $handle = $this->routes[$routingKey] ?? null;

        if ($handle === null) {
            $this->report("no handler for {$routingKey}, discarding message {$this->messageId($message)}");
            $message->ack();

            return;
        }

        try {
            $handle($message);
            $message->ack();
        } catch (PoisonMessageException $exception) {
            $this->report("message {$this->messageId($message)} dead-lettered without retrying: {$exception->getMessage()}");
            $message->nack();
        } catch (Throwable $exception) {
            $this->retryOrDeadLetter($message, $exception);
        }
    }

    private function retryOrDeadLetter(AMQPMessage $message, Throwable $exception): void
    {
        $attempts = $this->acquiredCount($message);
        $limit = (int) config('rabbitmq.consumer.retry_limit');

        if ($attempts >= $limit) {
            $this->report("message {$this->messageId($message)} dead-lettered after {$attempts} attempts: {$exception->getMessage()}");
            $message->nack();

            return;
        }

        $delay = $this->retryDelay($attempts);

        $this->report("message {$this->messageId($message)} returned to {$this->queue}, retrying in {$delay}s: {$exception->getMessage()}");
        ($this->sleep)($delay);
        $message->nack(requeue: true);
    }

    private function retryDelay(int $attempts): int
    {
        $base = (int) config('rabbitmq.consumer.retry_backoff_base_seconds');
        $cap = (int) config('rabbitmq.consumer.retry_backoff_cap_seconds');

        $delay = $base * (2 ** $attempts);

        return $delay > $cap || $delay <= 0 ? $cap : (int) $delay;
    }

    private function acquiredCount(AMQPMessage $message): int
    {
        $headers = $message->has('application_headers')
            ? $message->get('application_headers')->getNativeData()
            : [];

        foreach (['x-acquired-count', 'x-delivery-count'] as $header) {
            if (isset($headers[$header]) && is_numeric($headers[$header])) {
                return (int) $headers[$header];
            }
        }

        return 0;
    }

    private function messageId(AMQPMessage $message): string
    {
        return $message->has('message_id') ? (string) $message->get('message_id') : 'unknown';
    }

    private function report(string $line): void
    {
        ($this->log)($line);
    }
}

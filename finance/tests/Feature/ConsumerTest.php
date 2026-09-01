<?php

use App\Messaging\Consumer;
use App\Messaging\PoisonMessageException;
use App\Messaging\Topology;
use PhpAmqpLib\Message\AMQPMessage;
use RuntimeException;

beforeEach(function () {
    $this->routingKey = 'invoice.closed';
    $this->scratch = declareScratchTopology($this->routingKey);

    purgeQueue(Topology::DEAD_LETTER_QUEUE);
});

afterEach(function () {
    deleteScratchTopology($this->scratch['queue']);
});

function publishToScratch(string $exchange, string $routingKey, string $body = '{"invoiceId":42}'): void
{
    broker()->channel()->basic_publish(
        new AMQPMessage($body, ['message_id' => 'test-message', 'content_type' => 'application/json']),
        $exchange,
        $routingKey,
    );
}

/**
 * @param  array<string, callable(AMQPMessage): void>  $routes
 */
function consumer(string $queue, array $routes): Consumer
{
    return Consumer::for($queue, $routes, log: fn (string $line) => null, sleep: fn (int $seconds) => null);
}

it('routes a message to the handler of its key and acknowledges it', function () {
    $handled = null;

    publishToScratch($this->scratch['exchange'], $this->routingKey);

    consumer($this->scratch['queue'], [
        $this->routingKey => function (AMQPMessage $message) use (&$handled) {
            $handled = json_decode($message->getBody(), true);
        },
    ])->consume(stopAfter: 1, waitTimeout: 5);

    expect($handled)->toBe(['invoiceId' => 42])
        ->and(popMessage($this->scratch['queue'], timeout: 1.0))->toBeNull();
});

it('acknowledges and discards a key with no handler', function () {
    publishToScratch($this->scratch['exchange'], $this->routingKey);

    consumer($this->scratch['queue'], [])->consume(stopAfter: 1, waitTimeout: 5);

    expect(popMessage($this->scratch['queue'], timeout: 1.0))->toBeNull()
        ->and(popMessage(Topology::DEAD_LETTER_QUEUE, timeout: 1.0))->toBeNull();
});

it('dead-letters a poison message without retrying', function () {
    $attempts = 0;

    publishToScratch($this->scratch['exchange'], $this->routingKey, 'not json');

    consumer($this->scratch['queue'], [
        $this->routingKey => function (AMQPMessage $message) use (&$attempts) {
            $attempts++;

            throw new PoisonMessageException(new RuntimeException('invalid payload'));
        },
    ])->consume(stopAfter: 1, waitTimeout: 5);

    expect($attempts)->toBe(1)
        ->and(popMessage(Topology::DEAD_LETTER_QUEUE))->not->toBeNull();
});

it('retries a transient failure and dead-letters once the budget runs out', function () {
    config(['rabbitmq.consumer.retry_limit' => 1]);

    $attempts = 0;

    publishToScratch($this->scratch['exchange'], $this->routingKey);

    consumer($this->scratch['queue'], [
        $this->routingKey => function (AMQPMessage $message) use (&$attempts) {
            $attempts++;

            throw new RuntimeException('database is down');
        },
    ])->consume(stopAfter: 2, waitTimeout: 5);

    expect($attempts)->toBe(2)
        ->and(popMessage(Topology::DEAD_LETTER_QUEUE))->not->toBeNull()
        ->and(popMessage($this->scratch['queue'], timeout: 1.0))->toBeNull();
});

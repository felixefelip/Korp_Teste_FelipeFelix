<?php

use App\Messaging\Connection;
use App\Messaging\Topology;
use Illuminate\Foundation\Testing\RefreshDatabase;
use PhpAmqpLib\Message\AMQPMessage;
use PhpAmqpLib\Wire\AMQPTable;
use Tests\TestCase;

/*
|--------------------------------------------------------------------------
| Test Case
|--------------------------------------------------------------------------
|
| The closure you provide to your test functions is always bound to a specific PHPUnit test
| case class. By default, that class is "PHPUnit\Framework\TestCase". Of course, you may
| need to change it using the "pest()" function to bind different classes or traits.
|
*/

pest()->extend(TestCase::class)
    ->use(RefreshDatabase::class)
    ->in('Feature');

/*
|--------------------------------------------------------------------------
| Expectations
|--------------------------------------------------------------------------
|
| When you're writing tests, you often need to check that values meet certain conditions. The
| "expect()" function gives you access to a set of "expectations" methods that you can use
| to assert different things. Of course, you may extend the Expectation API at any time.
|
*/

expect()->extend('toBeOne', function () {
    return $this->toBe(1);
});

/*
|--------------------------------------------------------------------------
| Functions
|--------------------------------------------------------------------------
|
| While Pest is very powerful out-of-the-box, you may have some testing code specific to your
| project that you don't want to repeat in every file. Here you can also expose helpers as
| global functions to help you to reduce the number of lines of code in your test files.
|
*/

function broker(): Connection
{
    static $connection = null;

    if (! $connection instanceof Connection || ! $connection->isOpen()) {
        $connection = Connection::open();
    }

    return $connection;
}

function purgeQueue(string ...$queues): void
{
    foreach ($queues as $queue) {
        broker()->channel()->queue_purge($queue);
    }
}

function popMessage(string $queue, float $timeout = 5.0): ?AMQPMessage
{
    $deadline = microtime(true) + $timeout;

    do {
        $message = broker()->channel()->basic_get($queue);

        if ($message instanceof AMQPMessage) {
            $message->ack();

            return $message;
        }

        usleep(50_000);
    } while (microtime(true) < $deadline);

    return null;
}

/**
 * @return array{exchange: string, queue: string}
 */
function declareScratchTopology(string $routingKey): array
{
    $name = 'finance.test-'.bin2hex(random_bytes(6));
    $channel = broker()->channel();

    $channel->exchange_declare($name, 'topic', false, true, false);
    $channel->queue_declare($name, false, true, false, false, false, new AMQPTable([
        'x-queue-type' => 'quorum',
        'x-dead-letter-exchange' => Topology::DEAD_LETTER_EXCHANGE,
    ]));
    $channel->queue_bind($name, $name, $routingKey);

    return ['exchange' => $name, 'queue' => $name];
}

function deleteScratchTopology(string $name): void
{
    broker()->channel()->queue_delete($name);
    broker()->channel()->exchange_delete($name);
}

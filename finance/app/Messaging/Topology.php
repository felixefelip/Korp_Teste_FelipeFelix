<?php

namespace App\Messaging;

use PhpAmqpLib\Channel\AMQPChannel;
use PhpAmqpLib\Wire\AMQPTable;

class Topology
{
    public const FINANCE_EXCHANGE = 'finance.events';

    public const BILLING_EXCHANGE = 'billing.events';

    public const DEAD_LETTER_EXCHANGE = 'finance.dead-letter';

    public const INVOICE_EVENTS_QUEUE = 'finance.invoice-events';

    public const DEAD_LETTER_QUEUE = 'finance.dead-letter';

    public const INVOICE_CLOSED_KEY = 'invoice.closed';

    public const INVOICE_REOPENED_KEY = 'invoice.reopened';

    public static function declareOn(AMQPChannel $channel): void
    {
        $channel->exchange_declare(self::FINANCE_EXCHANGE, 'topic', false, true, false);
        $channel->exchange_declare(self::BILLING_EXCHANGE, 'topic', false, true, false);

        self::declareDeadLetter($channel);

        $channel->queue_declare(
            self::INVOICE_EVENTS_QUEUE, false, true, false, false, false, self::consumedQueueArguments()
        );

        self::bind($channel, self::INVOICE_EVENTS_QUEUE, self::BILLING_EXCHANGE, [
            self::INVOICE_CLOSED_KEY,
            self::INVOICE_REOPENED_KEY,
        ]);
    }

    private static function declareDeadLetter(AMQPChannel $channel): void
    {
        $channel->exchange_declare(self::DEAD_LETTER_EXCHANGE, 'topic', false, true, false);

        $channel->queue_declare(self::DEAD_LETTER_QUEUE, false, true, false, false, false, new AMQPTable([
            'x-queue-type' => 'quorum',
        ]));

        self::bind($channel, self::DEAD_LETTER_QUEUE, self::DEAD_LETTER_EXCHANGE, ['#']);
    }

    private static function consumedQueueArguments(): AMQPTable
    {
        return new AMQPTable([
            'x-queue-type' => 'quorum',
            'x-dead-letter-exchange' => self::DEAD_LETTER_EXCHANGE,
        ]);
    }

    /**
     * @param  list<string>  $routingKeys
     */
    private static function bind(AMQPChannel $channel, string $queue, string $exchange, array $routingKeys): void
    {
        foreach ($routingKeys as $routingKey) {
            $channel->queue_bind($queue, $exchange, $routingKey);
        }
    }
}

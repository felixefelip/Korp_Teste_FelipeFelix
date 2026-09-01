<?php

namespace App\Messaging;

use PhpAmqpLib\Channel\AMQPChannel;
use PhpAmqpLib\Connection\AMQPStreamConnection;
use PhpAmqpLib\Message\AMQPMessage;

class Connection
{
    public function __construct(
        private readonly AMQPStreamConnection $connection,
        private readonly AMQPChannel $channel,
    ) {}

    public static function open(): self
    {
        $url = parse_url((string) config('rabbitmq.url'));

        $connection = new AMQPStreamConnection(
            $url['host'] ?? 'rabbitmq',
            $url['port'] ?? 5672,
            $url['user'] ?? 'guest',
            $url['pass'] ?? 'guest',
            isset($url['path']) ? rawurldecode(ltrim($url['path'], '/')) ?: '/' : '/',
            connection_timeout: (float) config('rabbitmq.connect_timeout'),
            read_write_timeout: (float) config('rabbitmq.heartbeat') * 2,
            keepalive: true,
            heartbeat: (int) config('rabbitmq.heartbeat'),
        );

        $channel = $connection->channel();

        Topology::declareOn($channel);

        $channel->basic_qos(0, (int) config('rabbitmq.prefetch'), false);
        $channel->confirm_select();

        return new self($connection, $channel);
    }

    public function channel(): AMQPChannel
    {
        return $this->channel;
    }

    public function isOpen(): bool
    {
        return $this->connection->isConnected() && $this->channel->is_open();
    }

    public function close(): void
    {
        rescue(fn () => $this->channel->close(), report: false);
        rescue(fn () => $this->connection->close(), report: false);
    }

    public function publish(AMQPMessage $message, string $exchange, string $routingKey, float $timeout): void
    {
        $returned = null;
        $refused = false;

        $this->channel->set_return_listener(
            function (int $replyCode, string $replyText, string $returnExchange, string $returnKey) use (&$returned): void {
                $returned = $returnKey;
            }
        );

        $this->channel->set_nack_handler(function (AMQPMessage $nacked) use (&$refused): void {
            $refused = true;
        });

        $this->channel->basic_publish($message, $exchange, $routingKey, mandatory: true);
        $this->channel->wait_for_pending_acks_returns($timeout);

        if ($returned !== null) {
            throw new UnroutableMessageException($returned);
        }

        if ($refused) {
            throw new BrokerRefusedException($routingKey);
        }
    }
}

<?php

use App\Messaging\Connection;
use App\Messaging\Topology;
use PhpAmqpLib\Message\AMQPMessage;

it('declares the topology more than once without failing', function () {
    Connection::open()->close();
    Connection::open()->close();
})->throwsNoExceptions();

it('routes invoice events published by billing to the finance queue', function (string $routingKey) {
    purgeQueue(Topology::INVOICE_EVENTS_QUEUE);

    broker()->channel()->basic_publish(
        new AMQPMessage('{"invoiceId":42}', ['content_type' => 'application/json']),
        Topology::BILLING_EXCHANGE,
        $routingKey,
    );

    $message = popMessage(Topology::INVOICE_EVENTS_QUEUE);

    expect($message)->not->toBeNull()
        ->and($message->getRoutingKey())->toBe($routingKey)
        ->and($message->getBody())->toBe('{"invoiceId":42}');
})->with([
    Topology::INVOICE_CLOSED_KEY,
    Topology::INVOICE_REOPENED_KEY,
]);

it('does not route unbound keys to the finance queue', function () {
    purgeQueue(Topology::INVOICE_EVENTS_QUEUE);

    broker()->channel()->basic_publish(
        new AMQPMessage('{}'),
        Topology::BILLING_EXCHANGE,
        'invoice.close.requested',
    );

    expect(popMessage(Topology::INVOICE_EVENTS_QUEUE, timeout: 1.0))->toBeNull();
});

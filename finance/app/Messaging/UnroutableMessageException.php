<?php

namespace App\Messaging;

use RuntimeException;

class UnroutableMessageException extends RuntimeException
{
    public function __construct(string $routingKey)
    {
        parent::__construct('no queue bound to the routing key: '.$routingKey);
    }
}

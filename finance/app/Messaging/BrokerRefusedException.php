<?php

namespace App\Messaging;

use RuntimeException;

class BrokerRefusedException extends RuntimeException
{
    public function __construct(string $routingKey)
    {
        parent::__construct('broker refused the message: '.$routingKey);
    }
}

<?php

namespace App\Messaging;

use RuntimeException;
use Throwable;

class PoisonMessageException extends RuntimeException
{
    public function __construct(Throwable $cause)
    {
        parent::__construct('message will never be handled: '.$cause->getMessage(), 0, $cause);
    }
}

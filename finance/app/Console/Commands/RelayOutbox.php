<?php

namespace App\Console\Commands;

use App\Messaging\Relay;
use Illuminate\Console\Command;

class RelayOutbox extends Command
{
    protected $signature = 'finance:relay';

    protected $description = 'Publish pending outbox events to RabbitMQ';

    public function handle(): void
    {
        Relay::make(log: fn (string $line) => $this->line($line))->run();
    }
}

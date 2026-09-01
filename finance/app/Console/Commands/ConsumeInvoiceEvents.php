<?php

namespace App\Console\Commands;

use App\Messaging\Consumer;
use App\Messaging\Topology;
use Illuminate\Console\Command;
use PhpAmqpLib\Message\AMQPMessage;

class ConsumeInvoiceEvents extends Command
{
    protected $signature = 'finance:consume';

    protected $description = 'Consume invoice events published by billing';

    public function handle(): void
    {
        Consumer::for(
            Topology::INVOICE_EVENTS_QUEUE,
            $this->routes(),
            log: fn (string $line) => $this->line($line),
        )->run();
    }

    /**
     * @return array<string, callable(AMQPMessage): void>
     */
    private function routes(): array
    {
        return [];
    }
}

<?php

namespace App\Messaging;

class Signals
{
    private bool $stopping = false;

    public static function trap(): self
    {
        $signals = new self;

        if (! function_exists('pcntl_async_signals')) {
            return $signals;
        }

        pcntl_async_signals(true);

        foreach ([SIGTERM, SIGINT] as $signal) {
            pcntl_signal($signal, fn () => $signals->stop());
        }

        return $signals;
    }

    public function stop(): void
    {
        $this->stopping = true;
    }

    /**
     * @phpstan-impure
     */
    public function stopping(): bool
    {
        return $this->stopping;
    }
}

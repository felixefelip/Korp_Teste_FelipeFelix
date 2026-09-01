<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Collection;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

/**
 * @property int $id
 * @property string $event_id
 * @property string|null $causation_id
 * @property string $aggregate_type
 * @property int $aggregate_id
 * @property string $routing_key
 * @property array<string, mixed> $payload
 * @property Carbon $created_at
 * @property Carbon|null $published_at
 * @property int $attempts
 * @property Carbon $next_attempt_at
 * @property string|null $last_error
 */
class OutboxEvent extends Model
{
    public const AGGREGATE_INVOICE = 'invoice';

    public $timestamps = false;

    protected $guarded = [];

    /**
     * @return array<string, string>
     */
    protected function casts(): array
    {
        return [
            'payload' => 'array',
            'created_at' => 'datetime',
            'published_at' => 'datetime',
            'next_attempt_at' => 'datetime',
            'attempts' => 'integer',
        ];
    }

    /**
     * @param  array<string, mixed>  $payload
     */
    public static function record(
        string $aggregateType,
        int $aggregateId,
        string $routingKey,
        array $payload,
        ?string $causationId = null,
    ): self {
        $eventId = (string) Str::uuid();
        $occurredAt = Carbon::now();

        $envelope = array_filter([
            'eventId' => $eventId,
            'causationId' => $causationId,
        ], fn (?string $value): bool => $value !== null);

        return self::create([
            'event_id' => $eventId,
            'causation_id' => $causationId,
            'aggregate_type' => $aggregateType,
            'aggregate_id' => $aggregateId,
            'routing_key' => $routingKey,
            'payload' => [...$envelope, 'occurredAt' => $occurredAt->toIso8601ZuluString(), ...$payload],
            'created_at' => $occurredAt,
            'next_attempt_at' => $occurredAt,
        ]);
    }

    /**
     * @return Collection<int, self>
     */
    public static function claim(int $limit, int $leaseSeconds): Collection
    {
        return DB::transaction(function () use ($limit, $leaseSeconds): Collection {
            $events = self::query()
                ->whereNull('published_at')
                ->where('next_attempt_at', '<=', Carbon::now())
                ->orderBy('id')
                ->limit($limit)
                ->lock('for update skip locked')
                ->get();

            if ($events->isEmpty()) {
                return $events;
            }

            self::query()
                ->whereIn('id', $events->modelKeys())
                ->update(['next_attempt_at' => Carbon::now()->addSeconds($leaseSeconds)]);

            return $events;
        });
    }

    public function markPublished(): void
    {
        self::query()->whereKey($this->getKey())->update([
            'published_at' => Carbon::now(),
            'last_error' => null,
        ]);
    }

    public function markFailed(string $cause, Carbon $nextAttemptAt): void
    {
        self::query()->whereKey($this->getKey())->update([
            'attempts' => DB::raw('attempts + 1'),
            'last_error' => $cause,
            'next_attempt_at' => $nextAttemptAt,
        ]);
    }

    public function published(): bool
    {
        return $this->published_at !== null;
    }
}

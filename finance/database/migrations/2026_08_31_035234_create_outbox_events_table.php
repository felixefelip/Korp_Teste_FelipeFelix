<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('outbox_events', function (Blueprint $table) {
            $table->id();
            $table->uuid('event_id')->unique();
            $table->string('causation_id', 36)->nullable();
            $table->string('aggregate_type', 30);
            $table->unsignedBigInteger('aggregate_id')->index();
            $table->string('routing_key', 60);
            $table->jsonb('payload');
            $table->timestampTz('created_at');
            $table->timestampTz('published_at')->nullable();
            $table->unsignedInteger('attempts')->default(0);
            $table->timestampTz('next_attempt_at')->index();
            $table->text('last_error')->nullable();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('outbox_events');
    }
};

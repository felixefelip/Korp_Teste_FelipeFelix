<?php

return [

    'url' => env('RABBITMQ_URL', 'amqp://guest:guest@rabbitmq:5672/'),

    'connect_timeout' => (int) env('RABBITMQ_CONNECT_TIMEOUT', 5),

    'heartbeat' => (int) env('RABBITMQ_HEARTBEAT', 30),

    'prefetch' => (int) env('RABBITMQ_PREFETCH', 1),

    'consumer' => [
        'reconnect_seconds' => (int) env('RABBITMQ_RECONNECT_SECONDS', 3),
        'retry_limit' => (int) env('RABBITMQ_RETRY_LIMIT', 10),
        'retry_backoff_base_seconds' => (int) env('RABBITMQ_RETRY_BACKOFF_BASE', 2),
        'retry_backoff_cap_seconds' => (int) env('RABBITMQ_RETRY_BACKOFF_CAP', 30),
    ],

    'relay' => [
        'interval_seconds' => (int) env('RABBITMQ_RELAY_INTERVAL', 1),
        'batch' => (int) env('RABBITMQ_RELAY_BATCH', 50),
        'lease_seconds' => (int) env('RABBITMQ_RELAY_LEASE', 30),
        'publish_timeout_seconds' => (int) env('RABBITMQ_PUBLISH_TIMEOUT', 5),
        'backoff_base_seconds' => (int) env('RABBITMQ_RELAY_BACKOFF_BASE', 2),
        'backoff_cap_seconds' => (int) env('RABBITMQ_RELAY_BACKOFF_CAP', 60),
    ],

];

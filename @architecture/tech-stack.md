# Technical Stack

## Core Language and Runtime
- Go 1.22+ for native concurrency, module support, and security updates.

## HTTP Framework
- `github.com/gin-gonic/gin` for routing, middleware, error handling, and HTTP utilities.

## Queue Provider
- `github.com/rabbitmq/amqp091-go` to interact with RabbitMQ servers using AMQP 0-9-1.
- `github.com/cenkalti/backoff/v4` (or equivalent) for connection retry policies during startup and reconciliation.

## Data Access
- `github.com/jackc/pgx/v5` for PostgreSQL connectivity with efficient pooling and context-aware queries.
- `github.com/jackc/pgx/v5/pgxpool` to manage read-only connection pools.

## Configuration & Utilities
- `github.com/kelseyhightower/envconfig` (or similar) for environment-based configuration.
- `github.com/rs/zerolog` (or equivalent structured logger) for consistent observability.

These libraries are intentionally modular so that provider or database implementations can be swapped with limited surface area changes.



# Queue Manager Service

The queue manager is a Go backend responsible for provisioning and maintaining messaging infrastructure for a microservice ecosystem. It reconciles the desired topology defined in PostgreSQL with the actual state of the queue provider (initially RabbitMQ) and exposes operational endpoints for observability and manual synchronization.

## Key Responsibilities
- Reconcile exchanges, queues, and bindings at startup and on demand.
- Continuously verify messaging resources through scheduled health checks.
- Provide read-only access to expected topology definitions sourced from the database.
- Expose HTTP endpoints for health, readiness, and reconciliation insights.

## Documentation Map
- Architecture
  - [`@architecture/architecture.md`](./@architecture/architecture.md) — Four-layer system overview.
  - [`@architecture/queue-provider.md`](./@architecture/queue-provider.md) — Queue provider abstraction and RabbitMQ implementation notes.
  - [`@architecture/tech-stack.md`](./@architecture/tech-stack.md) — Go libraries and supporting components.
- Concepts
  - [`@concepts/overview.md`](./@concepts/overview.md) — Service intent and operating model.
- Data Model
  - [`@tables/README.md`](./@tables/README.md) — Definitions of queues, exchanges, service assignments, and bindings tables.
- API Surface
  - [`@apis/README.md`](./@apis/README.md) — Endpoint specs: requests, responses, examples.
  - [`@apis/response-format.md`](./@apis/response-format.md) — Standard response envelope (`message`, `data`, `metadata`).

## Runtime Behavior
1. Load expected state from PostgreSQL.
2. Connect to RabbitMQ via the provider abstraction.
3. Validate existing exchanges, queues, and bindings; create or remove resources to align with expectations.
4. Serve HTTP endpoints for health, readiness, and topology inspection.
5. Periodically re-run reconciliation to maintain alignment.

## Data Ownership
- PostgreSQL migrations are the sole mechanism for altering topology definitions.
- The repository layer exposes read-only accessors; runtime components never mutate database records.
- Health checks and status verification rely on real-time provider queries rather than cached data.

## Next Steps
- Implement the provider interface and repository contracts.
- Wire Gin controllers to expose the documented endpoints.
- Configure continuous reconciliation intervals and logging/metrics integration.



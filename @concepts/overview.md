# Queue Manager Concept

## Primary Responsibility
The service owns the life cycle of messaging infrastructure required by downstream microservices. It interprets the desired state stored in PostgreSQL tables and ensures the queue provider reflects that state.

## Startup Behavior
1. Connect to PostgreSQL and fetch the set of expected exchanges, queues, bindings, and service assignments.
2. Establish a connection with the queue provider (RabbitMQ).
3. Compare the expected topology with the provider's current state.
4. Create missing exchanges, queues, and bindings; remove obsolete resources that are no longer defined.

## Continuous Operations
- Expose health and readiness endpoints for platform monitoring.
- Periodically perform reconciliation health checks to verify RabbitMQ resources remain present and healthy.
- Allow operators to trigger an on-demand synchronization via the `/sync` endpoint.

## Data Ownership
- The service does not mutate database records at runtime.
- Migration scripts define and evolve the expected state. Runtime logic consumes this data in a read-only fashion to inform reconciliation.

## Provider Verification
- Health and status verification calls query RabbitMQ directly.
- No cached or in-memory prediction of provider status is used; real-time information is always sourced from the provider.

By centralizing queue topology management, the queue manager removes duplicated bootstrap logic from other services and provides an auditable, declarative source of truth.



# Architecture Overview

The queue manager is organized as a four-layer backend service. Each layer owns a narrow responsibility, allowing loose coupling and focused testing boundaries.

## Controller Layer
- Exposes HTTP routes and protocol-level concerns only (routing, marshaling, response codes).
- Delegates all business decisions to the application service layer.
- Built with `github.com/gin-gonic/gin` for routing and middleware orchestration.

## Application Service Layer
- Entry point for business logic initiated by controllers or scheduled jobs.
- Performs request-centric orchestration: validating inputs, coordinating service-layer calls, translating provider errors into domain responses.
- Maintains transaction boundaries where required, leveraging repository interfaces for data access and queue provider interfaces for runtime state.

## Service Layer
- Contains cohesive modules that encapsulate domain rules around queue, exchange, binding, and assignment management.
- Encodes service-centric behaviors such as reconciliation, health verification, and provider synchronization.
- Designed for reuse across multiple application service scenarios, keeping logic provider-agnostic via abstractions.

## Repository Layer
- Sole layer that talks to PostgreSQL.
- Provides read-only access methods to retrieve expected state definitions (queues, exchanges, service assignments, bindings).
- Executes SQL queries generated via migrations; no runtime INSERT/UPDATE/DELETE operations are exposed.
- Uses strongly typed DTOs to return deterministic snapshots of expected resources.



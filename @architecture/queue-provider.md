# Queue Provider Abstraction

## Provider Interface
The provider module standardizes interactions with any queuing backend. The abstraction exposes the following capabilities:

- `Connect(ctx context.Context) error`
- `Disconnect(ctx context.Context) error`
- `CheckQueue(name string) (QueueStatus, error)`
- `CheckQueues(names []string) (map[string]QueueStatus, error)`
- `CheckExchange(name string) (ExchangeStatus, error)`
- `CheckExchanges(names []string) (map[string]ExchangeStatus, error)`
- `CreateQueue(def QueueDefinition) error`
- `CreateExchange(def ExchangeDefinition) error`
- `DeleteQueue(name string) error`
- `DeleteExchange(name string) error`
- `CheckBinding(binding BindingDefinition) (BindingStatus, error)`
- `CreateBinding(binding BindingDefinition) error`
- `DeleteBinding(binding BindingDefinition) error`
- `Health(ctx context.Context) error`

Each method focuses on a single responsibility so that higher layers can orchestrate workflows such as reconciliation or health checks without duplicating provider-specific logic.

### Method Semantics and Expectations

#### Connect(ctx)
- Purpose: Establish a connection/session to the underlying provider.
- Behavior:
  - Idempotent: multiple calls either re-use an open connection or no-op.
  - Implements retry with backoff until `ctx` is done.
  - Must validate credentials/endpoint and fail fast for permanent errors.
- Errors: honor `ctx` cancellation/deadline; classify transient vs. permanent where possible.

#### Disconnect(ctx)
- Purpose: Gracefully close connections and free resources.
- Behavior:
  - Idempotent: safe to call when not connected.
  - Best-effort flush of pending ops if applicable; bounded by `ctx`.
- Errors: return error only for unexpected failures; tolerate “already closed”.

#### CheckQueue(name)
- Purpose: Retrieve the current state of a single queue.
- Returns: `QueueStatus` including at minimum:
  - `exists`: bool
  - `state`: enum/string (e.g., healthy/degraded/unknown)
  - `durable`, `autoDelete`: when discoverable
  - `consumers`, `messagesReady`, `messagesUnacked`: when supported by provider
- Behavior: Read-only; no mutations; consistent with provider visibility guarantees.

#### CheckQueues(names)
- Purpose: Batch variant of `CheckQueue` to reduce round-trips.
- Returns: map keyed by queue name; missing entries imply “not found” or are present with `exists=false`.
- Behavior: Partial failures should still return best-effort results with an aggregated error when appropriate.

#### CheckExchange(name)
- Purpose: Retrieve the current state of a single exchange/topic entity.
- Returns: `ExchangeStatus` including at minimum `exists`, `type` (if available), `state`.
- Behavior: Read-only; consistent with provider semantics.

#### CheckExchanges(names)
- Purpose: Batch variant of `CheckExchange`.
- Returns: map of name to `ExchangeStatus`, similar partial-failure semantics as `CheckQueues`.

#### CreateQueue(def)
- Purpose: Ensure a queue exists with the specified definition.
- Behavior:
  - Must be idempotent: declaring an existing queue with the same properties should succeed.
  - If properties conflict with an existing queue, must return a descriptive error indicating mismatch.
  - Should validate `def` (name, flags, arguments) before attempting creation.
- Errors: distinguish validation errors from provider/transport errors.

#### CreateExchange(def)
- Purpose: Ensure an exchange exists with the specified definition.
- Behavior:
  - Idempotent on re-declare with identical properties.
  - Property conflicts reported explicitly.
- Errors: validation vs. provider errors are differentiated where feasible.

#### DeleteQueue(name)
- Purpose: Remove a queue.
- Behavior:
  - If the provider supports conditional deletion (e.g., only-if-empty), the implementation should document the policy; otherwise attempt deletion and return provider error if not allowed.
  - No-op when the queue does not exist (return success) unless the provider cannot distinguish; in that case return a typed NotFound error.

#### DeleteExchange(name)
- Purpose: Remove an exchange.
- Behavior: Mirrors `DeleteQueue` semantics, including NotFound handling and provider-specific constraints.

#### CheckBinding(binding)
- Purpose: Verify presence and characteristics of a binding from exchange to queue (with optional routing key/arguments).
- Returns: `BindingStatus` including `exists` and, when applicable, the actual `routingKey`/`arguments` observed for comparison.
- Behavior: Read-only; must not create bindings.

#### CreateBinding(binding)
- Purpose: Ensure a specific binding exists.
- Behavior:
  - Idempotent: creating an identical binding should succeed with no changes.
  - If a binding exists but mismatches (e.g., different routing key), the implementation should return a mismatch error rather than silently altering it unless the provider supports an update primitive.

#### DeleteBinding(binding)
- Purpose: Remove a specific binding.
- Behavior:
  - No-op/success if binding is already absent where the provider returns NotFound.
  - Otherwise forward provider errors (e.g., permission issues).

#### Health(ctx)
- Purpose: Lightweight provider-level health signal for liveness/readiness integrations.
- Behavior:
  - Should avoid heavyweight operations; e.g., ping or minimal metadata query.
  - Must observe `ctx` timeout/cancellation.
- Returns: nil on healthy; error with cause detail on unhealthy.

## Types and Data Contracts
These types are referenced by the interface and should be satisfied by each provider implementation:

- `QueueDefinition`:
  - `name`: string
  - `durable`: bool
  - `autoDelete`: bool
  - `arguments`: map[string]any (provider-specific knobs)
  - `description`: string (optional, not enforced by all providers)

- `ExchangeDefinition`:
  - `name`: string
  - `type`: string (e.g., direct, topic, fanout, headers)
  - `durable`: bool
  - `autoDelete`: bool
  - `internal`: bool
  - `arguments`: map[string]any
  - `description`: string (optional)

- `BindingDefinition`:
  - `exchangeName`: string
  - `queueName`: string
  - `routingKey`: string
  - `arguments`: map[string]any (e.g., headers matchers)
  - `mandatory`: bool (if applicable)

- `QueueStatus`:
  - `exists`: bool
  - `state`: string (e.g., healthy, degraded, unknown)
  - `durable`, `autoDelete`: bool (when discoverable)
  - `consumers`: int (if supported)
  - `messagesReady`, `messagesUnacked`: int (if supported)

- `ExchangeStatus`:
  - `exists`: bool
  - `type`: string (if discoverable)
  - `state`: string

- `BindingStatus`:
  - `exists`: bool
  - `routingKey`: string (actual)
  - `arguments`: map[string]any (actual, if available)

### Error Handling and Observability
- All methods must honor `context.Context` for cancellation and deadlines.
- Prefer typed or wrapped errors to signal NotFound, Conflict, Validation, and Transient conditions.
- Implementations should emit:
  - Structured logs with operation, target, outcome, and latency.
  - Metrics (counters, histograms) for success/failure and latency per method.
  - Traces/spans when tracing is enabled in the host application.

## RabbitMQ Implementation
The initial implementation targets RabbitMQ and relies on `github.com/rabbitmq/amqp091-go` for AMQP operations. It handles:

- Connection lifecycle management with exponential backoff retries.
- Declaring exchanges and queues idempotently, respecting durable and auto-delete flags stored in expected definitions.
- Verifying bindings by inspecting the existing exchange-to-queue relationships.
- Ensuring publisher confirm mode where needed for accurate status checks.

Implementations for other providers (e.g., AWS SQS, Google Pub/Sub) can be added by satisfying the same interface, enabling runtime selection without modifying upper layers.

### RabbitMQ-specific Notes
- Connection:
  - Uses separate channels per concern (declare, inspect, publish) to avoid head-of-line blocking.
  - Recovers channels after transient failures; respects `context` on retries.
- Queues:
  - `CreateQueue` maps to `QueueDeclare` with durable/autoDelete/args as provided.
  - Mismatched declares (e.g., durable flag change) return a conflict error surfaced to callers.
- Exchanges:
  - `CreateExchange` maps to `ExchangeDeclare`; supports `direct|topic|fanout|headers`.
  - Internal exchanges are supported via the `internal` flag in `ExchangeDefinition`.
- Bindings:
  - `CreateBinding` uses `QueueBind` with routing key and arguments.
  - `CheckBinding` infers existence via `QueueInspect`/`Exchange` metadata when available or via management API if configured.
- Status:
  - `CheckQueue` and `CheckExchange` populate `exists` and best-effort health derived from message rates or policy matches when accessible.
  - Consumer and message depth metrics are included when management telemetry is enabled.



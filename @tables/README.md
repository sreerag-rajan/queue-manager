# Data Model

The service references four PostgreSQL tables that define the expected messaging topology. All tables are treated as read-only by the application; changes are introduced exclusively via SQL migration files maintained outside of runtime.

## Common Columns (present on every table)
| Column | Type | Notes |
| --- | --- | --- |
| `id` | `bigint` | Primary key, surrogate integer identifier. |
| `uuid` | `uuid` | Globally unique business identifier; unique constraint. |
| `created_at` | `timestamptz` | Insert timestamp (UTC). |
| `updated_at` | `timestamptz` | Last update timestamp (UTC). |
| `deleted_at` | `timestamptz` | Soft-delete timestamp (UTC); `NULL` means active. |
| `meta` | `jsonb` | Free-form metadata for operators/automation. |

Recommendations:
- Maintain `updated_at` via triggers.
- Use partial unique indexes on natural keys with `WHERE deleted_at IS NULL` to allow safe soft-deletes.
- Always index `uuid` (unique) for cross-system correlation.

## `queues`
| Column | Type | Notes |
| --- | --- | --- |
| `id` | `bigint` | Primary key. |
| `uuid` | `uuid` | Unique; alternative external identifier. |
| `created_at` | `timestamptz` |  |
| `updated_at` | `timestamptz` |  |
| `deleted_at` | `timestamptz` | DEFAULT `NULL` (soft delete). |
| `meta` | `jsonb` | Operator metadata. |
| `queue_name` | `text` | Primary key, human-readable name. |
| `durable` | `boolean` | Whether the queue survives broker restarts. |
| `auto_delete` | `boolean` | Indicates whether the queue should auto-delete when unused. |
| `arguments` | `jsonb` | Broker-specific arguments (e.g., dead-letter config). |
| `description` | `text` | Optional documentation for operators. |

- Represents every queue that must exist in the provider.
- Repository layer exposes read methods such as `ListQueues()` that return deterministic snapshots.

Indexes:
- `UNIQUE (uuid)`
- `UNIQUE (queue_name) WHERE deleted_at IS NULL` (enforces unique active names with soft-delete)
- `GIN (arguments jsonb_path_ops)`
- `GIN (meta)`

Foreign Keys and Joins:
- Referenced by: `bindings.queue_name` → `queues.queue_name`; `service_assignments.queue_name` → `queues.queue_name`.
- Typical joins:
  - `bindings` on `queue_name`
  - `service_assignments` on `queue_name`

## `exchanges`
| Column | Type | Notes |
| --- | --- | --- |
| `id` | `bigint` | Primary key. |
| `uuid` | `uuid` | Unique. |
| `created_at` | `timestamptz` |  |
| `updated_at` | `timestamptz` |  |
| `deleted_at` | `timestamptz` | DEFAULT `NULL`. |
| `meta` | `jsonb` | Operator metadata. |
| `exchange_name` | `text` | Primary key. |
| `exchange_type` | `text` | AMQP type such as `direct`, `topic`, `fanout`, or `headers`. |
| `durable` | `boolean` | Durable exchanges persist across restarts. |
| `auto_delete` | `boolean` | Auto-delete behavior flag. |
| `internal` | `boolean` | Marks exchanges hidden from publishers. |
| `arguments` | `jsonb` | Broker arguments (e.g., alternate exchange). |
| `description` | `text` | Operational notes. |

- Declares the canonical list of exchanges that must be present.
- Read-only repository method `ListExchanges()` returns these definitions.

Indexes:
- `UNIQUE (uuid)`
- `UNIQUE (exchange_name) WHERE deleted_at IS NULL`
- `GIN (arguments jsonb_path_ops)`
- `GIN (meta)`

Foreign Keys and Joins:
- Referenced by: `bindings.exchange_name` → `exchanges.exchange_name`.
- Typical joins:
  - `bindings` on `exchange_name`

## `service_assignments`
| Column | Type | Notes |
| --- | --- | --- |
| `id` | `bigint` | Primary key. |
| `uuid` | `uuid` | Unique. |
| `created_at` | `timestamptz` |  |
| `updated_at` | `timestamptz` |  |
| `deleted_at` | `timestamptz` | DEFAULT `NULL`. |
| `meta` | `jsonb` | Operator metadata. |
| `service_name` | `text` | Consumer service identifier. |
| `queue_name` | `text` | References `queues.queue_name`. |
| `prefetch_count` | `integer` | Suggested prefetch for bindings/consumers. |
| `max_inflight` | `integer` | Operational limit for concurrent deliveries. |
| `notes` | `text` | Additional context (e.g., ownership contact). |

- Composite primary key on `(service_name, queue_name)`.
- Captures which microservice consumes which queue and associated operational hints.

Indexes:
- `UNIQUE (uuid)`
- `UNIQUE (service_name, queue_name) WHERE deleted_at IS NULL`
- `BTREE (service_name)`
- `BTREE (queue_name)`
- `GIN (meta)`

Foreign Keys and Joins:
- `FOREIGN KEY (queue_name) REFERENCES queues(queue_name)`
- Typical joins:
  - Join `queues` on `service_assignments.queue_name = queues.queue_name`

## `bindings`
| Column | Type | Notes |
| --- | --- | --- |
| `id` | `bigint` | Primary key. |
| `uuid` | `uuid` | Unique. |
| `created_at` | `timestamptz` |  |
| `updated_at` | `timestamptz` |  |
| `deleted_at` | `timestamptz` | DEFAULT `NULL`. |
| `meta` | `jsonb` | Operator metadata. |
| `exchange_name` | `text` | References `exchanges.exchange_name`. |
| `queue_name` | `text` | References `queues.queue_name`. |
| `routing_key` | `text` | Pattern or specific routing key. |
| `arguments` | `jsonb` | Additional binding arguments (e.g., headers). |
| `mandatory` | `boolean` | Indicates if binding removal should trigger alerts. |

- Composite primary key on `(exchange_name, queue_name, routing_key)`.
- Defines how messages flow from exchanges to queues.

Indexes:
- `UNIQUE (uuid)`
- `UNIQUE (exchange_name, queue_name, routing_key) WHERE deleted_at IS NULL`
- `BTREE (exchange_name, routing_key)`
- `BTREE (queue_name)`
- `GIN (arguments jsonb_path_ops)`
- `GIN (meta)`

Foreign Keys and Joins:
- `FOREIGN KEY (exchange_name) REFERENCES exchanges(exchange_name)`
- `FOREIGN KEY (queue_name) REFERENCES queues(queue_name)`
- Typical joins:
  - Join `exchanges` on `bindings.exchange_name = exchanges.exchange_name`
  - Join `queues` on `bindings.queue_name = queues.queue_name`

## Operational Guidance
- Repository layer implements `List*` methods only; no insert/update/delete operations are exposed.
- Health verification queries the provider directly rather than the database.
- PostgreSQL acts as the source of truth for the desired state, while RabbitMQ is the runtime state that must be reconciled.
- Soft delete via `deleted_at`: application queries SHOULD filter `WHERE deleted_at IS NULL` for active records and rely on partial unique indexes on natural keys.
- Consider triggers to maintain `updated_at` and to enforce JSON schema in `arguments`/`meta` where necessary.



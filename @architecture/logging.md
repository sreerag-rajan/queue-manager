# Logging Standard

This document defines a structured, five-level logging standard for the Queue Manager service.

## Levels
- `trace`: Extremely verbose, step-by-step flow and intermediate values; disabled in production by default.
- `debug`: Developer diagnostics for non-critical troubleshooting; safe to enable in staging.
- `info`: Lifecycle and expected events (start, stop, health, sync completed).
- `warn`: Suspicious or recoverable anomalies; partial failures with fallback.
- `error`: Failures that require attention; operations did not succeed.

## Categories
- `SYSTEM`: Logs produced by internal components (server startup/shutdown, DB connect, queue provider connect, reconciliation jobs, health checks).
- `API`: Logs produced in the context of an HTTP request (handlers, validation, downstream calls).

## Log Schema (JSON)
All logs MUST be newline-delimited JSON and include the following fields:
- `timestamp` (string): RFC3339Nano timestamp.
- `level` (string): One of `trace|debug|info|warn|error`.
- `category` (string): One of `SYSTEM|API`.
- `service` (string): Constant `"queue-manager"`.
- `traceId` (string): Correlation ID. Required for API logs; optional for SYSTEM logs (preferred if operation has a trace context).
- `file` (string): Source file name (best effort).
- `line` (number): Source line number (best effort).
- `message` (string): Human-readable summary.
- `body` (object|null): Structured details; MUST avoid secrets/PII. Redact or omit sensitive values.
- `tags` (string[]): Free-form labels to ease searching and filtering (e.g., `["api:sync","provider:rabbitmq","op:reconcile"]`).

Optional context objects:
- `http` (object, API only):
  - `method` (string)
  - `path` (string)
  - `status` (number)
  - `durationMs` (number)
  - `clientIp` (string)
  - `userAgent` (string)
- `db` (object, SYSTEM/API when relevant): e.g., `{"op":"connect","dsn":"<redacted>"}`.
- `provider` (object, SYSTEM/API when relevant): e.g., `{"name":"rabbitmq","host":"mq:5672"}`.

## Examples

### SYSTEM (startup info)
```
{
  "timestamp": "2025-11-24T09:30:00.123456Z",
  "level": "info",
  "category": "SYSTEM",
  "service": "queue-manager",
  "traceId": "",
  "file": "cmd/server/main.go",
  "line": 42,
  "message": "HTTP server started",
  "body": { "addr": ":8080" },
  "tags": ["system:start","http:listen"]
}
```

### API (request handled, with trace)
```
{
  "timestamp": "2025-11-24T09:31:10.420Z",
  "level": "info",
  "category": "API",
  "service": "queue-manager",
  "traceId": "8a1f3d6a-9b4c-4fbc-9a8b-7b0c2f9b7b3d",
  "file": "internal/api/sync.go",
  "line": 87,
  "message": "manual sync completed",
  "body": { "summary": { "create": 1, "delete": 0, "fix": 2 } },
  "tags": ["api:sync","scope:orders"],
  "http": { "method": "POST", "path": "/sync", "status": 202, "durationMs": 152, "clientIp": "203.0.113.8", "userAgent": "curl/8.0.1" }
}
```

### WARN (provider transient)
```
{
  "timestamp": "2025-11-24T09:32:05.111Z",
  "level": "warn",
  "category": "SYSTEM",
  "service": "queue-manager",
  "traceId": "4ff16de3-1e48-4a50-8a88-7d9e8edb8a5d",
  "file": "internal/provider/rabbitmq/connect.go",
  "line": 133,
  "message": "provider connect transient failure, will retry",
  "body": { "attempt": 3, "backoffMs": 2000, "error": "connection reset by peer" },
  "tags": ["provider:rabbitmq","connect","retry"]
}
```

## Emission Rules
- Always emit JSON, one object per line.
- Always include `service`, `level`, `category`, `message`, and `tags`.
- For API logs, `traceId` is mandatory and must be provided by middleware via context.
- For SYSTEM logs, if no request context is present, either omit `traceId` or generate a scoped one per operation (preferred for long-running jobs/reconciliation).
- Do not log secrets (passwords, tokens, DSNs). Redact or drop fields.
- For high-volume paths, use `trace` (and optionally `debug`) sampling in production to control cost.

## Tagging Guidance
Use short, consistent tags:
- Domain tags: `provider:rabbitmq`, `db:postgres`, `api:sync`, `api:details`.
- Operation tags: `op:connect`, `op:reconcile`, `op:declare`, `op:bind`.
- Result tags: `result:success`, `result:partial`, `result:failure`.

These enable quick filtering and dashboards (e.g., count of `level=error` grouped by `tags.provider:*`).

## Correlation and Propagation
- Inbound HTTP requests carry/receive `X-Trace-Id`. Middleware ensures a UUIDv4 is present for all routes except `/health` and `/ready`.
- All logs in a request lifecycle MUST include the same `traceId`.
- Outbound calls (HTTP, DB, provider) SHOULD propagate the `X-Trace-Id` when supported. At minimum, include it in logs.
- For background jobs (no HTTP), generate a new `traceId` at job start and reuse it within the job scope.

## Level Usage Guide
- `trace`: enter/exit of functions, intermediate payloads (redacted), loop iterations. Use sampling.
- `debug`: decisive branches, cache misses, retries scheduled, configuration resolution (non-secret).
- `info`: startup/shutdown, readiness achieved, sync started/completed, routine successful operations.
- `warn`: degraded provider health, partial reconciliation, nearing thresholds, non-fatal validation issues.
- `error`: failed connect/declare/bind, request handling errors (4xx with server conditions, 5xx), unrecoverable reconciliation failure.

## Mapping to External Sinks
The schema is compatible with common collectors (ELK/OpenSearch, Loki, Datadog). Ensure timestamps are RFC3339 (UTC) and that `level` and `traceId` are indexed. Configure retention and scrubbing at the sink level.



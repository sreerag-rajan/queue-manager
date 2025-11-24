# API Surface

All endpoints are served by the Gin HTTP server. Responses MUST conform to the standard envelope:

```
{
  "message": string,
  "data": any,
  "metadata": object
}
```

See response envelope details in `@apis/response-format.md`.

---

## 1. Health Check
- **Method**: `GET`
- **Path**: `/health`
- **Purpose**: Liveness probe.

### Request
- Query: none
- Body: none

### Responses
- 200 OK
  - Envelope:
    ```
    {
      "message": "healthy",
      "data": { "process": "up" },
      "metadata": {}
    }
    ```
- 503 Service Unavailable
  - Envelope:
    ```
    {
      "message": "unhealthy",
      "data": null,
      "metadata": { "reason": "startup not complete" }
    }
    ```

---

## 2. Readiness Check
- **Method**: `GET`
- **Path**: `/ready`
- **Purpose**: Readiness probe indicating dependencies are available and initial reconciliation finished.

### Request
- Query: none
- Body: none

### Responses
- 200 OK
  - Envelope:
    ```
    {
      "message": "ready",
      "data": {
        "database": "connected",
        "provider": "connected",
        "initialReconciliation": "completed"
      },
      "metadata": {}
    }
    ```
- 503 Service Unavailable
  - Envelope:
    ```
    {
      "message": "not ready",
      "data": {
        "database": "connected|disconnected",
        "provider": "connected|disconnected",
        "initialReconciliation": "pending|failed"
      },
      "metadata": { "retryAfterSeconds": 15 }
    }
    ```

---

## 3. Expected State
- **Method**: `GET`
- **Path**: `/expectation`
- **Purpose**: Returns expected topology from PostgreSQL.

### Request
- Query (optional):
  - `service`: string — filter by consumer service name
  - `queue`: string — filter by queue name
  - `exchange`: string — filter by exchange name
  - `page`: int — pagination page (default 1)
  - `pageSize`: int — items per page (default 100)
- Body: none

### Responses
- 200 OK
  - Envelope (example):
    ```
    {
      "message": "ok",
      "data": {
        "queues": [
          { "queue_name": "q.orders", "durable": true, "auto_delete": false, "arguments": {}, "description": "" }
        ],
        "exchanges": [
          { "exchange_name": "ex.orders", "exchange_type": "topic", "durable": true, "auto_delete": false, "internal": false, "arguments": {}, "description": "" }
        ],
        "service_assignments": [
          { "service_name": "orders-api", "queue_name": "q.orders", "prefetch_count": 50, "max_inflight": 200, "notes": "" }
        ],
        "bindings": [
          { "exchange_name": "ex.orders", "queue_name": "q.orders", "routing_key": "order.*", "arguments": {}, "mandatory": true }
        ]
      },
      "metadata": { "page": 1, "pageSize": 100, "total": 4 }
    }
    ```

---

## 4. Actual State
- **Method**: `GET`
- **Path**: `/reality`
- **Purpose**: Live provider state from RabbitMQ.

### Request
- Query (optional): `queue`, `exchange`, `page`, `pageSize`
- Body: none

### Responses
- 200 OK
  - Envelope (example):
    ```
    {
      "message": "ok",
      "data": {
        "queues": [
          { "name": "q.orders", "exists": true, "state": "healthy", "consumers": 5, "messagesReady": 0, "messagesUnacked": 0 }
        ],
        "exchanges": [
          { "name": "ex.orders", "exists": true, "state": "healthy" }
        ],
        "bindings": [
          { "exchange": "ex.orders", "queue": "q.orders", "routing_key": "order.*", "exists": true, "state": "healthy" }
        ]
      },
      "metadata": { "page": 1, "pageSize": 100, "total": 3, "timestamp": "RFC3339" }
    }
    ```

---

## 5. Comparative Details
- **Method**: `GET`
- **Path**: `/details`
- **Purpose**: Diff between expectation and reality.

### Request
- Query (optional): `service`, `queue`, `exchange`, `page`, `pageSize`
- Body: none

### Responses
- 200 OK
  - Envelope (example):
    ```
    {
      "message": "ok",
      "data": {
        "missing": {
          "queues": ["q.payments"],
          "exchanges": [],
          "bindings": [
            { "exchange_name": "ex.orders", "queue_name": "q.orders", "routing_key": "order.created" }
          ]
        },
        "unexpected": {
          "queues": ["q.legacy"],
          "exchanges": [],
          "bindings": []
        },
        "mismatched": {
          "bindings": [
            {
              "expected": { "exchange_name": "ex.orders", "queue_name": "q.orders", "routing_key": "order.*" },
              "actual":   { "exchange": "ex.orders", "queue": "q.orders", "routing_key": "order.#" }
            }
          ]
        },
        "healthy": {
          "queues": ["q.orders"],
          "exchanges": ["ex.orders"],
          "bindings": ["ex.orders -> q.orders (order.*)"]
        }
      },
      "metadata": { "summary": { "missing": 2, "unexpected": 1, "mismatched": 1, "healthy": 3 } }
    }
    ```

---

## 6. Manual Synchronization
- **Method**: `POST`
- **Path**: `/sync`
- **Purpose**: Trigger reconciliation without restart.

### Request
- Query (optional):
  - `dryRun`: bool — if true, compute actions but do not change provider (default false)
  - `service`, `queue`, `exchange`: string — limit scope of sync
- Body (optional):
  ```
  {
    "dryRun": false,
    "scope": { "service": "orders-api", "queues": ["q.orders"], "exchanges": ["ex.orders"] }
  }
  ```

### Responses
- 202 Accepted (async sync started)
  - Envelope:
    ```
    {
      "message": "sync started",
      "data": { "jobId": "uuid", "startedAt": "RFC3339" },
      "metadata": {}
    }
    ```
- 200 OK (dry run)
  - Envelope:
    ```
    {
      "message": "dry run",
      "data": {
        "actions": {
          "toCreate": { "queues": ["q.payments"], "exchanges": [], "bindings": [] },
          "toDelete": { "queues": ["q.legacy"], "exchanges": [], "bindings": [] },
          "toFix":    { "bindings": ["ex.orders -> q.orders (order.# -> order.*)"] }
        }
      },
      "metadata": { "summary": { "create": 1, "delete": 1, "fix": 1 } }
    }
    ```
- 400 Bad Request
  - Envelope with validation errors:
    ```
    {
      "message": "invalid request",
      "data": null,
      "metadata": { "errors": [{ "field": "pageSize", "issue": "must be between 1 and 500" }] }
    }
    ```

---

### Common Considerations
- All endpoints must return the standard envelope.
- Pagination fields (`page`, `pageSize`, `total`) belong in `metadata`.
- Errors use `message` for a human-readable summary; `data` can be null; `metadata.errors` may include structured details.
- Authentication/authorization may be applied via middleware.
- Version/compatibility fields can be added to `metadata` as needed.
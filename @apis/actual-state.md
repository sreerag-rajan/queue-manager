# Actual State API

- Method: `GET`
- Path: `/reality`
- Purpose: Returns live provider state (e.g., RabbitMQ) for queues, exchanges, and bindings.

See the standard response envelope in `@apis/response-format.md`.

---

## Authentication
None by default. May be applied via middleware if configured.

---

## Request
- Path params: none
- Query params (all optional):
  - `queue`: string — filter by queue name
  - `exchange`: string — filter by exchange name
  - `page`: integer — pagination page (default: 1)
  - `pageSize`: integer — items per page (default: 100; valid range: 1–500)
- Headers: none required
- Body: none

---

## Responses

### 200 OK
- Meaning: Provider state returned successfully.
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

### 400 Bad Request
- Meaning: One or more query parameters are invalid.
- Envelope (example):
```
{
  "message": "invalid request",
  "data": null,
  "metadata": { "errors": [{ "field": "pageSize", "issue": "must be between 1 and 500" }] }
}
```

---

## Notes
- `metadata.timestamp` reports when the snapshot was generated.



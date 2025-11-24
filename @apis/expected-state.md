# Expected State API

- Method: `GET`
- Path: `/expectation`
- Purpose: Returns the expected messaging topology (queues, exchanges, bindings, service assignments) from PostgreSQL.

See the standard response envelope in `@apis/response-format.md`.

---

## Authentication
None by default. May be applied via middleware if configured.

---

## Request
- Path params: none
- Query params (all optional):
  - `service`: string — filter by consumer service name
  - `queue`: string — filter by queue name
  - `exchange`: string — filter by exchange name
  - `page`: integer — pagination page (default: 1)
  - `pageSize`: integer — items per page (default: 100; valid range: 1–500)
- Headers: none required
- Body: none

---

## Responses

### 200 OK
- Meaning: Expected topology returned successfully.
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
- Pagination fields (`page`, `pageSize`, `total`) are returned in `metadata`.



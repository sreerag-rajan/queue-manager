# Comparative Details API

- Method: `GET`
- Path: `/details`
- Purpose: Returns a diff between expectation (from PostgreSQL) and reality (from provider), grouped by missing, unexpected, mismatched, and healthy.

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
- Meaning: Comparison completed successfully.
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
- Intended for human-friendly diagnosis and automated reconciliation planning.



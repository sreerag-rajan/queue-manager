# Manual Synchronization API

- Method: `POST`
- Path: `/sync`
- Purpose: Triggers reconciliation of provider state with expectation without restarting the service. Supports dry-run for planning.

See the standard response envelope in `@apis/response-format.md`.

---

## Authentication
None by default. May be applied via middleware if configured.

---

## Request
- Path params: none
- Query params (all optional):
  - `dryRun`: boolean — if true, compute actions but do not change provider (default: false)
  - `service`: string — limit scope by service name
  - `queue`: string — limit scope by queue name
  - `exchange`: string — limit scope by exchange name
- Headers: `Content-Type: application/json` when sending a body
- Body (optional):
```
{
  "dryRun": false,
  "scope": { "service": "orders-api", "queues": ["q.orders"], "exchanges": ["ex.orders"] }
}
```

Notes:
- Query params and body may both be provided; body takes precedence if both specify the same field (implementation-specific; ensure consistent usage).
- Pagination does not apply.

---

## Responses

### 202 Accepted
- Meaning: A reconciliation job has been accepted and started asynchronously.
- Envelope:
```
{
  "message": "sync started",
  "data": { "jobId": "uuid", "startedAt": "RFC3339" },
  "metadata": {}
}
```

### 200 OK (Dry Run)
- Meaning: Dry run; no provider changes were made. Proposed actions are returned.
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

### 400 Bad Request
- Meaning: Invalid request (e.g., invalid types/values or conflicting parameters).
- Envelope:
```
{
  "message": "invalid request",
  "data": null,
  "metadata": { "errors": [{ "field": "pageSize", "issue": "must be between 1 and 500" }] }
}
```

---

## Notes
- For synchronous progress tracking and completion, use the returned `jobId` with the relevant job/status endpoint if available (out of scope here).
- Use dry run in CI or pre-deploy checks to preview changes safely.



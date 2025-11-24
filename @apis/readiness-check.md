# Readiness Check API

- Method: `GET`
- Path: `/ready`
- Purpose: Readiness probe indicating dependencies are available and initial reconciliation has finished.

See the standard response envelope in `@apis/response-format.md`.

---

## Authentication
None by default. May be applied via middleware if configured.

---

## Request
- Path params: none
- Query params: none
- Headers: none required
- Body: none

---

## Responses

### 200 OK
- Meaning: All required dependencies are connected and initial reconciliation completed.
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

### 503 Service Unavailable
- Meaning: One or more dependencies are unavailable or initial reconciliation is pending/failed.
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

## Notes
- Intended for orchestrator readiness checks and traffic gating.
- Unlike `/health`, this checks external dependencies and startup tasks.



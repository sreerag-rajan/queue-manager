# Health Check API

- Method: `GET`
- Path: `/health`
- Purpose: Liveness probe to indicate the process is up.

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
- Meaning: The service process is running and responsive.
- Envelope:
```
{
  "message": "healthy",
  "data": { "process": "up" },
  "metadata": {}
}
```

### 503 Service Unavailable
- Meaning: The service is not yet ready to report healthy (e.g., startup not complete) or a fatal condition is detected.
- Envelope:
```
{
  "message": "unhealthy",
  "data": null,
  "metadata": { "reason": "startup not complete" }
}
```

---

## Notes
- Intended for container/orchestrator liveness checks.
- Returns quickly without checking external dependencies.



# Middleware: Trace ID Generation

This document describes a single HTTP middleware that ensures every API request carries a correlation ID (`traceId`) used across logs and downstream calls.

## Goals
- Provide a stable, unique `traceId` (UUIDv4) for all API requests.
- Skip generating/attaching `traceId` for health endpoints: `/health` and `/ready`.
- Propagate the `traceId` through request context for logging and downstream calls.
- Echo the `traceId` back to clients in the response headers.

## Behavior
1. On each request:
   - If the path is `/health` or `/ready`, do nothing (no `traceId` set or echoed).
   - Else:
     - Read incoming `X-Trace-Id` header. If present and valid (UUIDv4 preferred), use it.
     - If missing or invalid, generate a new UUIDv4.
     - Store the `traceId` in the request context.
     - Set `X-Trace-Id` response header to the chosen value.
2. Handlers and logging components retrieve the `traceId` from context and include it in all `API` logs (see `@architecture/logging.md`).
3. Outbound calls (HTTP to other services) SHOULD forward `X-Trace-Id` to maintain end-to-end correlation.

## Header Conventions
- Request header: `X-Trace-Id`
- Response header: `X-Trace-Id`
- The middleware MUST NOT set these headers for `/health` or `/ready`.

## Context Key
Implementations SHOULD define a dedicated, unexported context key type to avoid collisions, e.g.:
- `type traceKey struct{}`
- `func WithTraceID(ctx context.Context, id string) context.Context`
- `func TraceIDFromContext(ctx context.Context) (string, bool)`

## Validation
- Accept only UUIDv4 (canonical, lowercase preferred). If a different format is received, either:
  - Normalize (lowercase, remove braces), and if still invalid
  - Generate a new UUIDv4.

## Pseudocode (Gin-style)
```
func TraceIDMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    p := c.FullPath()
    if p == "/health" || p == "/ready" {
      c.Next()
      return
    }
    id := c.GetHeader("X-Trace-Id")
    if !isValidUUIDv4(id) {
      id = uuid.NewString()
    }
    ctx := WithTraceID(c.Request.Context(), id)
    c.Request = c.Request.WithContext(ctx)
    c.Header("X-Trace-Id", id)
    c.Next()
  }
}
```

## Interaction with Logging
- The logger MUST pull `traceId` from context and inject it into all `API` logs.
- Ensure middleware runs before any request logging or other middlewares that may emit logs.
- For background/system tasks, generate and bind a `traceId` at the job boundary (see `@architecture/logging.md`).

## Exclusions
- `/health` and `/ready` requests SHOULD NOT carry `traceId` to keep probes lightweight and avoid noisy correlation.
- If a `X-Trace-Id` is provided for these endpoints, implementations MAY ignore it and omit the response header.

## Testing Considerations
- Requests without `X-Trace-Id` receive a UUIDv4 in response.
- Requests with a valid `X-Trace-Id` echo the same value.
- `/health` and `/ready` return no `X-Trace-Id`.
- Handlers can retrieve the `traceId` via context and logs contain the same ID.



# Standard Response Envelope

All HTTP responses MUST use the following envelope:

```
{
  "message": string,
  "data": any,
  "metadata": object
}
```

## Fields
- `message` — Human-readable summary for operators (short, actionable).
- `data` — Primary response payload. For errors, this may be `null`.
- `metadata` — Auxiliary data not core to the primary payload:
  - Pagination: `page`, `pageSize`, `total`
  - Timestamps: `timestamp` (RFC3339)
  - Filters: any query/body filters echoed back
  - Errors: `errors: [{ field, issue }]` for validation problems
  - Versioning: `apiVersion`, `schemaVersion` if needed

## Examples

### Success (paged list)
```
{
  "message": "ok",
  "data": [{ "id": "..." }],
  "metadata": { "page": 1, "pageSize": 50, "total": 123 }
}
```

### Validation Error
```
{
  "message": "invalid request",
  "data": null,
  "metadata": { "errors": [{ "field": "pageSize", "issue": "must be between 1 and 500" }] }
}
```

### Server Error
```
{
  "message": "internal error",
  "data": null,
  "metadata": { "requestId": "..." }
}
```



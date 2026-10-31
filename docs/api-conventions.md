
# PulseQueue API Conventions

## Base URL
Local development: http://localhost:8080

## Content types
- JSON requests must use Content-Type: application/json.
- JSON responses use Content-Type: application/json.

## Successful responses
Return an appropriate HTTP 2xx status and a documented
JSON response where a response body is expected.

## Error responses
All API errors use the following structure:

{
  "error": {
    "code": "stable_machine_readable_code",
    "message": "Human-readable explanation."
  }
}

## Request validation
- Reject malformed JSON.
- Reject unknown fields where strict request schemas apply.
- Reject unexpected trailing JSON values.
- Limit request body size.
- Validate input before invoking business logic.

## Security
- Never expose internal exception details or secrets.
- Do not log passwords, access tokens, or sensitive payloads.
- Use appropriate HTTP status codes.

## Versioning
Keep the initial API unversioned while the contract is
small. Introduce a versioning strategy before making
breaking changes to a published API.

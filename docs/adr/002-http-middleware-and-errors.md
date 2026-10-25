
# ADR 002: Standardize HTTP Middleware and API Errors

## Status
Accepted

## Context
PulseQueue needs consistent request logging and predictable
JSON errors for frontend and API clients.

## Decision
- Implement request logging as HTTP middleware.
- Use a shared JSON error response format.
- Return JSON 404 and 405 responses.
- Use Go's standard library for HTTP behavior.

## Consequences
- API responses are easier to consume and test.
- Request logging is centralized.
- Middleware can be reused across endpoints.
- Optional ResponseWriter interfaces may need explicit
  support if streaming or connection upgrades are introduced.

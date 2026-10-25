
# ADR 001: Use the Go Standard Library for HTTP

## Status
Accepted

## Context
PulseQueue is a learning-focused background job engine.
Understanding Go server-side development is a primary objective.

## Decision
Use net/http for routing and HTTP server implementation.
Use encoding/json for JSON responses.
Use log/slog for structured logging.

## Consequences
- Learn core Go HTTP concepts without framework abstractions.
- Keep the initial dependency list small.
- Implement and test middleware explicitly when needed.
- Reassess a third-party router only when requirements justify it.

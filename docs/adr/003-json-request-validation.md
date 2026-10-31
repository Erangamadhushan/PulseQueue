# ADR 003: Strict JSON Request Decoding

## Status

Accepted

## Context

PulseQueue will expose JSON APIs that accept client input.
Malformed or oversized requests must be rejected safely.

## Decision

Use encoding/json with strict field checking.
Limit JSON request bodies to 1 MiB by default.
Reject trailing JSON values and unsupported media types.
Separate decoding errors from business validation.

## Consequences

- Request behavior is predictable and testable.
- Unexpected JSON fields are rejected.
- Request-size limits reduce exposure to oversized bodies.
- Endpoint handlers remain responsible for mapping
  validation errors to appropriate HTTP responses.

# ADR 008: At-Least-Once Execution

## Status

Accepted

## Context

Handler side effects and the database status update cannot generally be
committed as one atomic operation. A worker can complete an external action
and then crash before persisting the completed state.

## Decision

Design PulseQueue for at-least-once execution. A job may execute again after a
worker crash, lease recovery, or equivalent interruption. Handler authors are
responsible for making external side effects idempotent where possible.

## Consequences

- The system prioritizes durable job progress over an unsupported
  exactly-once guarantee.
- Handlers should use idempotency keys or transactional deduplication for
  externally visible operations.
- Repeated execution must be documented and tested by handler authors.
- Retry policies and backoff remain separate decisions for a later stage.

# ADR 004: Separate Worker Process

## Status

Accepted

## Context

PulseQueue must execute background jobs without allowing job execution to
consume the API server's request-handling resources. Worker workloads may have
different scaling, shutdown, and operational requirements than the HTTP API.

## Decision

Run workers as a separate Go process from the API server. The API process
creates, lists, retrieves, and cancels jobs. The worker process claims and
executes jobs, then persists their results through the shared PostgreSQL
database.

## Consequences

- API and worker processes can be deployed and scaled independently.
- Worker shutdown and execution limits do not directly interrupt API requests.
- Both processes share database contracts and configuration conventions.
- Local development requires starting the worker separately when execution is
  needed.

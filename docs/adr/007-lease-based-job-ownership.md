# ADR 007: Lease-Based Job Ownership

## Status

Accepted

## Context

A worker may crash or lose connectivity after claiming a job. The database
must record which worker owns a running job and provide a basis for future
recovery of abandoned work.

## Decision

Assign each claimed job a worker ID, locked timestamp, and lease expiration
timestamp. Completion and failure updates must verify the current worker ID
and running status before changing the job. Lease renewal and expired-lease
recovery will be added before production use of long-running jobs.

## Consequences

- Ownership is explicit and completion cannot silently overwrite another
  worker's result.
- Lease metadata supports future crash recovery and worker monitoring.
- A lease alone does not recover abandoned jobs until recovery is implemented.
- Handlers must not assume that a lease provides exactly-once execution.

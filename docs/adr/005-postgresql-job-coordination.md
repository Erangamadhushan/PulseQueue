# ADR 005: PostgreSQL Job Coordination

## Status

Accepted

## Context

Multiple worker processes may poll for the same pending jobs. The system needs
an atomic claim operation that prevents two workers from owning one job at the
same time without introducing another infrastructure dependency.

## Decision

Use PostgreSQL as the job coordination mechanism. Claiming selects an eligible
pending job with `FOR UPDATE SKIP LOCKED`, updates its ownership and running
state in the same transaction, and commits before handler execution begins.

## Consequences

- Concurrent workers can claim different jobs safely.
- PostgreSQL remains the source of truth for job state and ownership.
- Claim transactions remain short because handlers run after commit.
- Worker availability depends on PostgreSQL availability.

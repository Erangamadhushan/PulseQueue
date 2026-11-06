# ADR 006: Registered Go Handlers

## Status

Accepted

## Context

Jobs store a type name, but execution requires mapping that type to application
code. The mapping should be explicit, type-safe, and easy to configure within a
worker process.

## Decision

Use an in-process registry of Go handlers. A handler has the signature
`func(context.Context, jobs.Job) error` and is registered by a non-empty job
type name. The registry rejects nil handlers and duplicate names and returns a
clear error for unknown types.

## Consequences

- Handler registration is simple and compile-time discoverable.
- Context cancellation can propagate into application handlers.
- A worker must be restarted when its handler set changes.
- Dynamic plugins and remote handler execution are outside the current scope.

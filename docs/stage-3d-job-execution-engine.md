# Stage 3D — Job Execution Engine

**Project:** PulseQueue  
**Milestone:** Stage 3 — Background Worker Engine  
**Phase:** 3D — Job Execution Engine  
**Status:** Proposed  
**Prerequisites:** Stage 3A, Stage 3B and Stage 3C  
**Backend:** Go  
**Database:** PostgreSQL  
**Driver:** pgx/v5

---

# 1. Introduction

## 1.1 Purpose

Stage 3D introduces the core job execution engine of PulseQueue.

The execution engine connects the PostgreSQL job claiming mechanism with the Go handler registry.

Its responsibility is to receive a claimed job, resolve its registered handler, execute that handler safely, and persist the execution result.

This phase transforms PulseQueue from a system that can assign jobs to workers into a system that can actually execute background jobs.

## 1.2 Current System

Completed phases provide:

- PostgreSQL persistence.
- Atomic job claiming with `FOR UPDATE SKIP LOCKED`.
- Worker ownership and claim tokens.
- Lease expiration information.
- Worker configuration.
- Queue subscriptions.
- Handler registration and lookup.
- Worker process bootstrap.

The missing component is the execution engine.

## 1.3 Main Objectives

1. Implement a reusable job execution component.
2. Integrate the Stage 3C handler registry.
3. Execute handlers using Go contexts.
4. Handle successful execution.
5. Handle returned errors.
6. Recover handler panics.
7. Persist completion and failure states.
8. Validate job ownership using claim tokens.
9. Implement lease renewal during execution.
10. Prevent stale workers from updating job results.
11. Add comprehensive unit and PostgreSQL integration tests.

## 1.4 Out of Scope

- Automatic retry strategies and exponential backoff.
- Multiple concurrent jobs per worker.
- Continuous polling.
- Worker scheduling and polling loops.
- Full graceful shutdown orchestration.
- Dead-letter queues.
- Distributed worker monitoring.

These will be addressed in later phases.

---

# 2. Execution Architecture

## 2.1 Component Overview

The execution engine will depend on:

- Job domain model.
- Handler registry.
- Job repository.
- Worker configuration.
- Logger.
- Execution context.

```text
             Claimed Job
                  |
                  v
           Execution Engine
                  |
          +-------+-------+
          |               |
          v               v
    Handler Registry   Repository
          |               |
          v               v
     Job Handler      PostgreSQL
          |               |
          +-------+-------+
                  |
             Execution Result
                  |
          +-------+-------+
          |               |
       Success           Failure
          |               |
          v               v
      Completed          Failed
```

## 2.2 Execution Responsibilities

The execution engine is responsible for:

1. Receiving a claimed job.
2. Validating its ownership information.
3. Resolving the handler.
4. Creating an execution context.
5. Starting lease renewal.
6. Running the handler.
7. Handling returned errors or panics.
8. Stopping lease renewal.
9. Persisting the final result using ownership validation.

The execution engine must not perform job claiming or continuous polling. Those responsibilities belong to the worker polling component.

---

# 3. Execution Contract

## 3.1 Claimed Job

The execution engine must receive both the claimed job and its claim token.

The existing Stage 3B `ClaimResult` should be reused.

Illustrative structure:

```go
type ClaimResult struct {
    Job        jobs.Job
    ClaimToken uuid.UUID
}
```

Avoid introducing another claim representation unless the existing implementation requires it.

## 3.2 Executor Interface

A possible executor contract:

```go
type Executor interface {
    Execute(
        ctx context.Context,
        claim jobs.ClaimResult,
    ) error
}
```

The exact package ownership should follow the existing project structure.

The executor should receive its dependencies through a constructor rather than creating global database or registry instances.

## 3.3 Execution Dependencies

Recommended dependencies:

- Job repository.
- Handler registry.
- Lease duration.
- Lease renewal interval.
- Logger.

A constructor can receive these dependencies.

Avoid coupling the executor directly to the HTTP API or frontend.

---

# 4. Handler Resolution

The executor uses the Stage 3C registry to resolve the job handler.

Example:

```go
handler, ok := registry.Get(claim.Job.Type)

if !ok {
    // Handle unknown handler.
}
```

The handler is responsible for application-specific business logic.

The executor is responsible for job lifecycle management.

## 4.1 Unknown Handler

If no handler is registered for the job type:

- Do not execute the job.
- Record a descriptive error.
- Mark the job failed using ownership-checked persistence.
- Clear ownership fields.

Unknown handler errors are terminal in V1.

Automatic retry and configurable error classification are deferred to Stage 4.

---

# 5. Execution Context

## 5.1 Purpose

Each handler must execute with a context that allows cancellation.

The execution context should derive from the parent worker context.

```go
executionCtx, cancel := context.WithCancel(parentCtx)
defer cancel()
```

This allows the worker to cancel execution when:

- The application is shutting down.
- The parent context is cancelled.
- The worker loses ownership of the job.
- A future execution deadline is reached.

## 5.2 Handler Contract

Reuse the Stage 3C handler contract:

```go
type Handler func(
    ctx context.Context,
    job jobs.Job,
) error
```

Handlers should respect context cancellation.

A handler that ignores its context may continue performing external operations after the worker has lost ownership.

Go cannot forcibly terminate an arbitrary handler goroutine. Context cancellation is cooperative.

---

# 6. Execution Result Handling

## 6.1 Successful Execution

When the handler returns `nil`:

1. Stop lease renewal.
2. Verify that the execution still owns the job.
3. Persist `completed`.
4. Set `completed_at`.
5. Update `updated_at`.
6. Clear ownership fields.

Expected state:

```text
status = completed
worker_id = NULL
locked_at = NULL
lease_expires_at = NULL
claim_token = NULL
```

## 6.2 Handler Failure

When the handler returns an error:

1. Stop lease renewal.
2. Verify current ownership.
3. Persist `failed`.
4. Store the execution error in `last_error`.
5. Set `updated_at`.
6. Clear ownership fields.

Expected state:

```text
status = failed
last_error = execution error
worker_id = NULL
locked_at = NULL
lease_expires_at = NULL
claim_token = NULL
```

Automatic retries are intentionally excluded from this phase.

## 6.3 Completion Repository Method

A suggested repository method:

```go
Complete(
    ctx context.Context,
    jobID uuid.UUID,
    claimToken uuid.UUID,
) error
```

## 6.4 Failure Repository Method

A suggested repository method:

```go
Fail(
    ctx context.Context,
    jobID uuid.UUID,
    claimToken uuid.UUID,
    executionErr error,
) error
```

Both operations must verify:

- The job ID matches.
- The job is still `running`.
- The claim token matches the current ownership token.

Both must clear all ownership fields when changing the job state.

## 6.5 Ownership Conflict

If an update affects zero rows, the worker no longer owns the job or the job is no longer in the expected state.

Return a recognizable ownership conflict error.

Do not report the operation as successful.

The worker must not attempt to overwrite the state using an unconditional update.

---

# 7. Panic Recovery

## 7.1 Why Panic Recovery?

A registered handler may panic because of:

- Unexpected nil pointer access.
- Invalid assumptions.
- Runtime errors.
- Third-party library behavior.

An unrecovered panic could terminate the worker process.

The executor should isolate panic recovery around the handler invocation.

## 7.2 Recovery Behavior

A panic should be converted into an execution error.

Example:

```go
func invokeHandler(
    ctx context.Context,
    handler Handler,
    job jobs.Job,
) (err error) {
    defer func() {
        if recovered := recover(); recovered != nil {
            err = fmt.Errorf(
                "handler panic: %v",
                recovered,
            )
        }
    }()

    return handler(ctx, job)
}
```

Recovery must occur in the same goroutine where the handler runs.

The execution engine should also record useful diagnostic information in its logs.

Avoid exposing sensitive panic details to public API responses.

In V1, a recovered panic results in a failed job.

---

# 8. Lease Renewal

## 8.1 Purpose

The job claiming phase introduced lease expiration.

A job can take longer to execute than its initial lease duration.

Without lease renewal, another worker could recover the job while the original handler is still running.

Therefore, the execution engine must renew ownership periodically.

## 8.2 Renewal Contract

A suggested repository method:

```go
RenewLease(
    ctx context.Context,
    jobID uuid.UUID,
    claimToken uuid.UUID,
    leaseDuration time.Duration,
) error
```

The update must verify the current claim token.

Example:

```sql
UPDATE jobs
SET
    lease_expires_at = NOW() + $3::interval,
    updated_at = NOW()
WHERE id = $1
  AND status = 'running'
  AND claim_token = $2;
```

If zero rows are affected, the worker no longer owns the job.

## 8.3 Renewal Interval

The renewal interval should be configurable or derived from the lease duration.

A reasonable initial policy is:

- Renew approximately every one-third of the lease duration.
- Leave sufficient time for database latency and temporary failures.
- Validate that the renewal interval is positive and shorter than the lease duration.

For example:

```text
Lease duration:       60 seconds
Renewal interval:     20 seconds
```

These values are initial defaults, not universal production settings.

## 8.4 Lease Renewal Failure

If lease renewal fails because ownership is lost:

1. Cancel the execution context.
2. Stop further renewal attempts.
3. Do not attempt an unconditional completion or failure update.
4. Return an ownership-related error.
5. Log the incident.

For transient database failures, use a bounded renewal strategy that leaves sufficient time before lease expiration.

Do not continue executing indefinitely when ownership cannot be confirmed.

## 8.5 Renewal and Completion Coordination

The executor must coordinate the renewal goroutine and handler result.

A safe sequence:

1. Start the renewal loop.
2. Execute the handler.
3. When the handler finishes, stop the renewal loop.
4. Wait for the renewal loop to exit.
5. Persist the result using the claim token.

This avoids racing a lease renewal against completion.

If the handler finishes at approximately the same time as ownership is lost, ownership validation remains authoritative.

---

# 9. Execution Lifecycle

The complete execution lifecycle:

```text
Receive ClaimResult
        |
        v
Validate Claim Information
        |
        v
Resolve Registered Handler
        |
        v
Create Execution Context
        |
        v
Start Lease Renewal
        |
        v
Execute Handler
        |
        +-------------------+
        |                   |
        v                   v
     Success              Error
        |                   |
        +---------+---------+
                  |
                  v
          Stop Lease Renewal
                  |
                  v
         Verify Current Ownership
                  |
          +-------+-------+
          |               |
          v               v
      Ownership OK    Ownership Lost
          |               |
          v               v
    Persist Result    Cancel/Return
          |
          v
      Finish Job
```

Panic recovery is part of the handler invocation boundary.

Unknown handlers follow the failure path without starting the handler.

---

# 10. Concurrency Model

## 10.1 V1 Execution

Each worker executes one job at a time.

This keeps the first execution engine simple.

Multiple independent worker processes can still execute different jobs concurrently.

Example:

```text
Worker A -> Job A
Worker B -> Job B
Worker C -> Job C
```

Within each worker process, only one job is executed at a time.

## 10.2 Future Concurrency

A future version may support multiple concurrent jobs per worker.

That will require:

- A bounded execution pool.
- Concurrency limits.
- Coordinated shutdown.
- Independent execution contexts.
- Independent lease renewal loops.
- Resource limits and backpressure.

Do not introduce this complexity in Stage 3D.

---

# 11. Error Handling

The execution engine must distinguish between handler errors, ownership errors and infrastructure errors.

| Error | Expected behavior |
|---|---|
| Handler returns error | Attempt to mark job failed |
| Handler panics | Recover and attempt to mark job failed |
| Unknown handler | Attempt to mark job failed |
| Ownership lost | Cancel handler and reject result update |
| Lease renewal fails | Cancel handler and report the failure |
| PostgreSQL unavailable | Return persistence error |
| Parent context cancelled | Stop execution and follow cancellation policy |
| Completion update fails | Return persistence error; do not report completion |
| Failure update fails | Return persistence error and preserve diagnostic context |

Automatic retry classification is deferred to Stage 4.

A persistence failure must not be confused with a successful handler execution.

---

# 12. Suggested Project Structure

```text
internal/
├── jobs/
│   ├── job.go
│   ├── repository.go
│   ├── postgres_repository.go
│   └── ...
│
└── worker/
    ├── config.go
    ├── handler.go
    ├── registry.go
    ├── executor.go
    ├── lease_renewer.go
    ├── errors.go
    └── ...
```

The exact file split should follow actual implementation complexity.

Avoid creating separate abstractions for every small function.

The repository owns database persistence. The executor owns the execution lifecycle.

---

# 13. Testing Strategy

## 13.1 Executor Unit Tests

Test:

- Successful handler execution.
- Handler error handling.
- Panic recovery.
- Unknown handler behavior.
- Invalid claim information.
- Context cancellation.
- Ownership loss handling.
- Completion persistence.
- Failure persistence.

Use fake repositories where appropriate to isolate execution behavior.

## 13.2 Lease Renewal Tests

Test:

- Renewal happens at the expected interval.
- Renewal extends the lease.
- Renewal uses the current claim token.
- Ownership loss cancels execution.
- Renewal goroutine stops after handler completion.
- Renewal does not race final result persistence.
- Parent cancellation stops renewal.

Avoid relying on long real-time sleeps. Use injected clocks or short controlled intervals where practical.

## 13.3 PostgreSQL Integration Tests

Using real PostgreSQL, verify:

### Completion

- Running job becomes completed.
- Completion timestamp is populated.
- Ownership fields are cleared.
- Incorrect claim token cannot complete the job.

### Failure

- Running job becomes failed.
- Last error is persisted.
- Ownership fields are cleared.
- Incorrect claim token cannot fail the job.

### Lease Renewal

- Lease expiry is extended.
- Incorrect claim token cannot renew the lease.
- Non-running jobs cannot renew their leases.

## 13.4 Reliability Tests

Test scenarios involving:

- Handler panic.
- Handler cancellation.
- Database failure during completion.
- Database failure during renewal.
- Ownership loss during execution.
- Handler completion during renewal.

The tests must confirm that stale workers cannot overwrite newer ownership states.

---

# 14. Implementation Checkpoints

| ID | Task | Expected Outcome |
|---|---|---|
| 3D.1 | Define execution contracts | Executor and dependencies established |
| 3D.2 | Extend repository lifecycle methods | Complete, Fail and RenewLease supported |
| 3D.3 | Implement ownership-checked SQL | Stale ownership rejected |
| 3D.4 | Implement handler invocation | Registered handlers execute |
| 3D.5 | Add panic recovery | Handler panic isolated |
| 3D.6 | Implement lease renewal | Active jobs maintain ownership |
| 3D.7 | Coordinate renewal and execution | Safe execution lifecycle |
| 3D.8 | Add unit and integration tests | Execution behavior verified |
| 3D.9 | Update documentation | Implementation recorded |
| 3D.10 | Run validation and commit | Stage 3D completed |

---

# 15. Acceptance Criteria

## Execution

- [x] Executor receives a valid claim.
- [x] Handler is resolved through the registry.
- [x] Handler executes with a cancellable context.
- [x] Successful execution is persisted as completed.
- [x] Handler errors are persisted as failed.
- [x] Unknown handlers are handled.
- [x] Panics are recovered.

## Ownership

- [x] Completion verifies the claim token.
- [x] Failure verifies the claim token.
- [x] Lease renewal verifies the claim token.
- [x] Stale workers cannot update newer ownership states.
- [x] Ownership fields are cleared after completion or failure.

## Lease Renewal

- [x] Lease renewal runs while a handler executes.
- [x] Renewal uses the correct job and claim token.
- [x] Renewal interval is shorter than the lease duration.
- [x] Renewal stops after execution.
- [x] Ownership loss cancels execution.
- [x] Renewal and final result persistence are coordinated.

## Validation

- [x] Unit tests pass.
- [ ] PostgreSQL integration tests pass.
- [ ] Race detector passes.
- [x] `go vet ./...` passes.
- [x] Documentation is updated.
- [ ] Changes are committed cleanly.

---

# 16. Risks and Technical Considerations

## 16.1 At-Least-Once Execution

PulseQueue provides at-least-once execution semantics.

A worker may perform an external operation and crash before recording the result.

The job may later execute again.

Handlers must be idempotent wherever possible.

## 16.2 Context Cancellation Is Cooperative

Go cannot forcibly terminate a running function.

Handlers must respect context cancellation.

A handler that ignores cancellation may continue running after losing ownership.

## 16.3 External Side Effects

Claim tokens protect database state, not external operations.

An email, payment or API request might already have happened before ownership is lost.

Use idempotency keys and appropriate external deduplication mechanisms.

## 16.4 Database Failures

If PostgreSQL becomes unavailable, the worker may be unable to renew a lease or persist an execution result.

The executor must stop claiming success when persistence cannot be confirmed.

## 16.5 Panic Recovery

Panic recovery protects the worker process from a handler panic, but it cannot guarantee that external side effects were not partially performed.

Such jobs may require careful idempotency handling.

---

# 17. Future Development

| Phase | Focus |
|---|---|
| Stage 3E | Worker polling loop and continuous job processing |
| Stage 3F | Graceful shutdown and worker lifecycle |
| Stage 4 | Automatic retries, exponential backoff and error classification |
| Stage 5 | Abandoned-job recovery, heartbeat and reliability |
| Stage 6 | Worker concurrency, scaling and monitoring |

---

# 18. Final Summary

Stage 3D introduces the core execution engine for PulseQueue.

It connects the job claiming mechanism to the handler registry and provides safe execution, panic recovery, ownership-checked persistence and lease renewal.

The most important design principle is that a worker must never update a job simply because it previously claimed it.

Every lifecycle update must verify current ownership.

After completing Stage 3D, PulseQueue will have the core capabilities to claim and execute jobs safely, while the next phase will introduce continuous polling and job processing.

**Stage 3D outcome:** A tested Go execution engine capable of running registered job handlers, maintaining job ownership and safely persisting execution results.

**Current implementation status:** The execution path persists successful jobs as
`completed` and handler failures as `failed`. Failed jobs retain a null
`completed_at`; all result and lease updates require the current worker and
claim token. Unit tests and static analysis pass. PostgreSQL integration tests,
race-detector validation, and committing the milestone require an environment
with PostgreSQL and a CGO-enabled C compiler.

**Implementation status:** The worker execution path now coordinates handler
invocation, panic conversion, ownership-checked completion/failure persistence,
and periodic lease renewal. Renewal failures cancel the handler context and
prevent stale result updates. PostgreSQL lease renewal is guarded by the job
ID, worker ID, running status, and claim token.
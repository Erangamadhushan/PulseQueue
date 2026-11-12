# Stage 3E — Worker Polling Loop and Job Processing

**Project:** PulseQueue  
**Milestone:** Stage 3 — Background Worker Engine  
**Phase:** 3E — Worker Polling Loop and Job Processing  
**Status:** Proposed  
**Prerequisites:** Stage 3A, Stage 3B, Stage 3C and Stage 3D  
**Backend:** Go  
**Database:** PostgreSQL

---

# 1. Introduction

## 1.1 Purpose

Stage 3E introduces the continuous worker polling loop.

The polling loop connects the PostgreSQL job claiming mechanism with the job execution engine.

Its responsibility is to continuously:

1. Request the next eligible job.
2. Claim the job from PostgreSQL.
3. Pass the claimed job to the execution engine.
4. Wait when no job is available.
5. Handle temporary claiming errors.
6. Repeat until the worker is stopped.

This phase makes the existing worker components operate together as a complete background job processing loop.

## 1.2 Current System

Previous stages provide:

- PostgreSQL job persistence.
- Atomic job claiming.
- Worker ownership and claim tokens.
- Lease renewal.
- Worker configuration.
- Queue subscriptions.
- Handler registry.
- Job execution engine.
- Ownership-checked completion and failure.
- Panic recovery.

The missing component is the continuous processing coordinator.

## 1.3 Objectives

1. Implement a continuous worker polling loop.
2. Integrate the existing job repository.
3. Integrate the existing job execution engine.
4. Respect configured queue subscriptions.
5. Handle empty queues efficiently.
6. Handle temporary database errors.
7. Support context cancellation.
8. Ensure sequential job execution per worker.
9. Prevent unnecessary database polling.
10. Add comprehensive unit and integration tests.

## 1.4 Out of Scope

- Multiple concurrent jobs per worker.
- Dynamic concurrency adjustment.
- Signal handling and complete graceful shutdown.
- Worker heartbeat and monitoring endpoints.
- Automatic retry strategies.
- Dead-letter queues.
- Distributed scheduling.
- Advanced adaptive polling.

These features are reserved for later phases.

---

# 2. High-Level Architecture

## 2.1 Component Overview

The worker polling loop coordinates two existing components:

- Job claiming repository.
- Job execution engine.

```text
                 Worker Process
                       |
                       v
               Polling Loop
                       |
              +--------+--------+
              |                 |
              v                 v
        Job Repository     Job Executor
              |                 |
              v                 v
          PostgreSQL       Handler Registry
                                |
                                v
                          Job Handler
```

The polling loop should not implement database queries or handler execution directly.

It coordinates existing components through their interfaces.

## 2.2 Processing Lifecycle

```text
Worker Starts
      |
      v
Start Polling Loop
      |
      v
Check Context
      |
      v
Claim Next Job
      |
      v
Job Available?
   /       \
  No       Yes
  |         |
  v         v
Wait      Execute Job
  |         |
  |         v
  |      Execution Finished
  |         |
  +---------+
      |
      v
Next Poll
```

The process repeats until the worker context is cancelled or a non-recoverable worker-level error occurs.

---

# 3. Worker Poller Design

## 3.1 Responsibilities

The polling component is responsible for:

- Requesting eligible jobs.
- Passing worker ID, queues and lease duration to the repository.
- Executing successfully claimed jobs.
- Waiting when the queue is empty.
- Handling temporary claim failures.
- Respecting context cancellation.
- Logging relevant processing information.

The poller must not duplicate functionality already implemented in the repository or executor.

## 3.2 Poller Structure

A possible structure:

```go
type Poller struct {
    repository jobs.Repository
    executor   Executor
    config     WorkerConfig
    logger     *slog.Logger
}
```

Use the actual repository and executor interfaces already established in the project.

Avoid creating duplicate interfaces if existing interfaces already provide the required behavior.

## 3.3 Constructor

A suggested constructor:

```go
func NewPoller(
    repository jobs.Repository,
    executor Executor,
    config WorkerConfig,
    logger *slog.Logger,
) *Poller
```

The constructor should receive its dependencies rather than creating global database connections or registry instances.

Validate required dependencies and configuration either in the constructor or at the worker bootstrap boundary.

---

# 4. Claiming Jobs

## 4.1 Claim Input

The poller should construct the existing Stage 3B claim input.

Example:

```go
claimInput := jobs.ClaimJobInput{
    WorkerID:      config.ID,
    Queues:        config.Queues,
    LeaseDuration: config.LeaseDuration,
}
```

The poller must not generate its own claim token.

Claim token generation belongs to the repository's atomic claiming operation.

## 4.2 Claim Result

The repository returns either:

- A claimed job with its ownership token.
- No available job.
- An error.

Expected behavior:

| Result | Poller behavior |
|---|---|
| Valid claim | Execute the job |
| No available job | Wait and poll again |
| Temporary database error | Log and retry after a delay |
| Context cancellation | Stop the polling loop |

## 4.3 Ownership

The poller must pass the complete claim result to the executor.

It must not reconstruct a claim using only the job ID.

The claim token is required for safe execution and ownership-checked persistence.

---

# 5. Empty Queue Handling

## 5.1 Why Waiting Is Necessary

A worker may frequently find no available jobs.

Immediately repeating the database query can create unnecessary database load.

Instead, the poller should wait for the configured polling interval.

## 5.2 Polling Interval

The existing worker configuration provides:

```go
PollInterval time.Duration
```

Example:

```text
Poll interval: 2 seconds

10:00:00 - No jobs
10:00:02 - No jobs
10:00:04 - Job available
10:00:04 - Job claimed
```

The actual polling interval should be configurable.

## 5.3 Context-Aware Waiting

Use a timer or ticker with a context-aware select.

Illustrative helper:

```go
func waitForNextPoll(
    ctx context.Context,
    interval time.Duration,
) error {
    timer := time.NewTimer(interval)
    defer timer.Stop()

    select {
    case <-ctx.Done():
        return ctx.Err()

    case <-timer.C:
        return nil
    }
}
```

The wait must stop immediately when the worker context is cancelled.

Avoid unconditional `time.Sleep`, because it cannot respond immediately to cancellation.

---

# 6. Database Error Handling

## 6.1 Temporary Errors

PostgreSQL may temporarily become unavailable because of:

- Network interruptions.
- Database restarts.
- Connection pool exhaustion.
- Temporary connection failures.
- Database failover.

A temporary claim error should not automatically terminate the worker.

The poller should log the error and retry after a bounded delay.

## 6.2 Polling Error Policy

For V1:

- Return or log the original error.
- Wait before attempting another claim.
- Use a configurable or internally defined bounded retry delay.
- Respect context cancellation during the wait.

Do not implement an aggressive immediate retry loop.

## 6.3 Bounded Backoff

A simple initial policy can use a fixed delay or capped exponential backoff.

Example:

```text
First error:  1 second
Second error: 2 seconds
Third error:  4 seconds
Maximum:      10 seconds
```

Reset the error backoff after a successful claim or a successful no-job result.

Automatic job retries are different from polling retries. Polling retries must not change a job's attempt count.

---

# 7. Job Execution

## 7.1 Sequential Execution

For V1, each worker processes one job at a time.

Example:

```text
Worker A:

Claim Job 1
     |
Execute Job 1
     |
Finish Job 1
     |
Claim Job 2
     |
Execute Job 2
```

This avoids introducing a worker execution pool before the basic lifecycle is validated.

Multiple worker processes can still execute different jobs concurrently.

## 7.2 Executor Integration

After a successful claim:

```go
err := executor.Execute(ctx, *claim)
```

The execution engine owns:

- Handler resolution.
- Handler invocation.
- Panic recovery.
- Lease renewal.
- Completion persistence.
- Failure persistence.
- Ownership validation.

The poller must not duplicate those responsibilities.

## 7.3 Execution Errors

An execution error should not automatically stop the polling loop.

For example:

```text
Job A -> Failed
Job B -> Completed
Job C -> Completed
```

The executor should report execution errors, and the poller should continue processing if the worker context remains active.

However, ownership loss and infrastructure errors should be logged with sufficient context.

The executor's error contract must distinguish expected job failures from conditions that make continued processing unsafe.

---

# 8. Polling Loop

## 8.1 Main Loop

A simplified example:

```go
func (p *Poller) Run(ctx context.Context) error {
    for {
        if err := ctx.Err(); err != nil {
            return err
        }

        claim, err := p.repository.ClaimNext(
            ctx,
            p.claimInput(),
        )

        if err != nil {
            if ctx.Err() != nil {
                return ctx.Err()
            }

            // Log and wait using a bounded retry policy.
            continue
        }

        if claim == nil {
            // No available job.
            // Wait for the next polling interval.
            continue
        }

        if err := p.executor.Execute(ctx, *claim); err != nil {
            // Record/log execution outcome.
            // Continue unless the context is cancelled
            // or the error makes further processing unsafe.
        }
    }
}
```

This illustrates the responsibilities and control flow.

The final implementation must include context-aware waits in both the empty-queue and database-error paths.

It must also avoid a busy loop after an execution error.

## 8.2 Run Contract

A possible interface:

```go
type Runner interface {
    Run(ctx context.Context) error
}
```

A context cancellation should result in a clean return according to the project's error conventions.

Do not hide non-context worker infrastructure errors.

---

# 9. Polling Behavior and Timing

## 9.1 Empty Queue

When no job is available:

1. Wait for the configured polling interval.
2. Check whether the context is cancelled.
3. Attempt another claim.

## 9.2 Successful Claim

When a job is available:

1. Execute immediately.
2. Wait for execution to finish.
3. Attempt the next claim.

Do not add an unnecessary polling delay after a successful execution.

## 9.3 Claim Failure

When the repository returns a temporary error:

1. Log the failure.
2. Wait using the bounded error backoff.
3. Retry the claim.

Do not change the job state or attempt count because a claim request failed.

---

# 10. Observability

The polling loop should provide structured logs.

Recommended events:

| Event | Suggested fields |
|---|---|
| Worker started | worker_id, queues |
| Job claimed | worker_id, job_id, job_type, queue |
| Job execution started | worker_id, job_id |
| Job execution finished | worker_id, job_id, duration |
| Job execution failed | worker_id, job_id, error |
| No job available | worker_id |
| Claim failed | worker_id, error |
| Polling stopped | worker_id |

Avoid logging sensitive job payloads.

Avoid logging at high volume for every empty polling attempt in production. Debug-level logging or sampled logs may be more appropriate.

Metrics can be introduced in Stage 6.

---

# 11. Error Handling

| Scenario | Expected behavior |
|---|---|
| No job available | Wait and poll again |
| Job successfully claimed | Execute immediately |
| Claim operation fails | Log and retry after backoff |
| Handler returns error | Executor records failure; continue polling |
| Handler panics | Executor recovers; continue polling |
| Context cancelled | Stop polling |
| Database remains unavailable | Retry with bounded delay |
| Executor reports ownership loss | Do not overwrite job; log incident |
| Invalid worker configuration | Fail during startup |

The poller must distinguish between job-level errors and worker-level errors.

A failed job should not automatically terminate a healthy worker.

---

# 12. Worker Bootstrap Integration

The worker bootstrap will connect the poller to existing infrastructure.

Startup sequence:

1. Load shared application configuration.
2. Load worker configuration.
3. Initialize logging.
4. Initialize PostgreSQL connection pool.
5. Verify database readiness.
6. Create handler registry.
7. Register application handlers.
8. Create executor.
9. Create poller.
10. Run the worker.

The worker should reuse existing database and logging infrastructure.

Full signal handling and graceful shutdown orchestration will be finalized in Stage 3F.

---

# 13. Testing Strategy

## 13.1 Poller Unit Tests

Use fake repositories and executors.

Test:

- Poller construction.
- Successful job claim.
- Successful execution.
- No available jobs.
- Repeated polling.
- Repository error handling.
- Execution error handling.
- Context cancellation.
- Empty queue waiting.
- Error backoff.
- No busy loop.

## 13.2 Repository Integration Tests

Reuse the existing PostgreSQL integration tests from Stage 3B.

Verify that the poller:

- Claims only subscribed queues.
- Processes scheduled jobs when eligible.
- Receives the claim token.
- Does not bypass atomic claiming.

## 13.3 Execution Integration Tests

Verify:

- Claimed jobs reach the executor.
- Completed jobs become completed.
- Failed jobs become failed.
- Ownership fields are cleared correctly.
- One worker does not execute multiple jobs concurrently.

## 13.4 Multiple Worker Tests

Run multiple pollers with separate worker IDs against the same PostgreSQL database.

Verify:

- Different jobs can be claimed by different workers.
- No simultaneous duplicate claim occurs.
- Each worker processes its subscribed queues.
- Job ownership remains consistent.

## 13.5 Cancellation Tests

Test cancellation during:

- Empty queue waiting.
- Database claiming.
- Error backoff.
- Job execution.

The poller must not claim another job after cancellation is observed.

Execution cancellation is cooperative and depends on handlers respecting their contexts.

---

# 14. Suggested Project Structure

```text
internal/
├── jobs/
│   └── ...
│
└── worker/
    ├── config.go
    ├── handler.go
    ├── registry.go
    ├── executor.go
    ├── lease_renewer.go
    ├── poller.go
    ├── poller_test.go
    └── ...
```

Keep the polling loop in the existing worker package unless the implementation becomes large enough to justify a separate package.

Avoid creating unnecessary interfaces or abstractions.

---

# 15. Implementation Checkpoints

| ID | Task | Expected Outcome |
|---|---|---|
| 3E.1 | Define poller contract | Poller dependencies established |
| 3E.2 | Implement polling loop | Continuous claim cycle |
| 3E.3 | Integrate repository | Jobs claimed using existing claim mechanism |
| 3E.4 | Integrate executor | Claimed jobs executed |
| 3E.5 | Implement empty queue waiting | No unnecessary database polling |
| 3E.6 | Implement error backoff | Temporary failures handled |
| 3E.7 | Implement context cancellation | Polling responds to cancellation |
| 3E.8 | Add unit and integration tests | Processing behavior verified |
| 3E.9 | Integrate worker bootstrap | Worker process runs poller |
| 3E.10 | Update documentation and commit | Stage 3E completed |

---

# 16. Acceptance Criteria

## Polling

- [x] Poller continuously requests eligible jobs.
- [x] Configured queues are respected.
- [x] Claim results preserve claim tokens.
- [x] Empty queues use context-aware waiting.
- [x] Successful claims execute immediately.
- [x] No busy loop occurs.

## Execution

- [x] Existing executor is reused.
- [x] Jobs are processed sequentially per worker.
- [x] Job failures do not terminate a healthy worker.
- [x] Ownership validation remains inside the execution and repository layers.

## Error Handling

- [x] Temporary claim errors are logged.
- [x] Bounded backoff is implemented.
- [x] Polling backoff does not modify job attempts.
- [x] Context cancellation interrupts waits.
- [x] Worker-level errors are not silently swallowed.

## Integration

- [x] Worker bootstrap initializes the poller.
- [x] Existing database initialization is reused.
- [x] Multiple worker processes work concurrently.
- [x] Queue subscriptions are respected.

## Validation

- [x] Unit tests pass.
- [x] PostgreSQL integration tests pass.
- [x] Concurrency tests pass.
- [ ] Race detector passes.
- [x] `go vet ./...` passes.
- [x] Documentation is updated.
- [ ] Changes are committed cleanly.

---

# 17. Risks and Technical Considerations

## 17.1 Database Polling Load

Very short polling intervals may create unnecessary database load.

The interval should be configurable and adjusted based on measured workload.

## 17.2 Busy Loops

Missing waits after empty results or temporary errors can cause CPU and database usage to increase rapidly.

Every non-processing path must have an appropriate wait or termination path.

## 17.3 Error Classification

Job execution errors and worker infrastructure errors are different.

The poller must not stop simply because an individual job fails.

However, persistent infrastructure problems should be visible and handled with bounded backoff.

## 17.4 Context Cancellation

The polling loop must stop accepting new jobs after cancellation.

A currently executing job may require additional shutdown handling. Full orchestration is covered in Stage 3F.

## 17.5 Sequential Processing

Sequential processing is intentional for V1.

Increasing worker concurrency will require independent execution and lease management for each in-flight job.

---

# 18. Future Development

| Phase | Focus |
|---|---|
| Stage 3F | Graceful shutdown, signal handling and resource lifecycle |
| Stage 4 | Automatic retry and backoff strategies |
| Stage 5 | Abandoned-job recovery and reliability |
| Stage 6 | Worker concurrency, scaling and monitoring |

---

# 19. Final Summary

Stage 3E connects the job claiming mechanism and execution engine into a continuously running worker process.

The poller coordinates job claiming, execution, empty queue waiting, temporary error handling and cancellation.

The design deliberately keeps the worker sequential in V1.

Multiple independent workers can still process different jobs concurrently through PostgreSQL coordination.

After completing Stage 3E, PulseQueue will have an operational continuous background job processing loop.

**Stage 3E outcome:** A tested Go worker polling loop that continuously claims and executes jobs while respecting queue subscriptions, ownership rules and context cancellation.
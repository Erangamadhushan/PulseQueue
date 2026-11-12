# PulseQueue – Worker Engine Architecture

**Document:** Worker Engine Architecture  
**Project:** PulseQueue  
**Milestone:** Stage 3  
**Status:** Implemented foundation – claim, registration, execution, completion and shutdown loop  
**Database:** PostgreSQL  
**Backend:** Go

---

## 1. Introduction

PulseQueue is a PostgreSQL-backed background job processing engine written in Go.

The existing system supports:

- Persistent job creation.
- Job retrieval and listing.
- Queue and status filtering.
- Pagination.
- Pending job cancellation.
- React-based job management dashboard.

Stage 3 now adds the worker process foundation, including PostgreSQL ownership
leases, registered handlers, execution failure isolation and context-driven
shutdown. Lease renewal, abandoned-lease recovery and retries remain follow-up
reliability work.

The next development milestone introduces a worker execution engine.

The worker engine will claim pending jobs from PostgreSQL, execute registered Go handlers and persist execution results.

The primary goal is to create a reliable, extensible and developer-friendly background job processing system.

## 2. Architecture Goals

The initial worker engine should:

1. Execute jobs stored in PostgreSQL.
2. Support registered Go handlers.
3. Prevent concurrent workers from claiming the same job simultaneously.
4. Track job ownership.
5. Handle execution errors and panics.
6. Support graceful worker shutdown.
7. Provide a foundation for retries and concurrent execution.
8. Keep the architecture simple enough to understand and maintain.

### Non-goals for the initial implementation

The first worker implementation will not include:

- Distributed scheduling.
- Automatic retry policies.
- Multiple priority levels.
- Dynamic plugin loading.
- Workflow orchestration.
- Exactly-once execution guarantees.
- Distributed event streaming.

These capabilities can be considered in later milestones.

## 3. High-Level Architecture

The API server and worker engine will run as separate Go processes.

### API Server

Responsibilities:

- Create jobs.
- Retrieve jobs.
- List and filter jobs.
- Cancel pending jobs.
- Expose health and readiness endpoints.

### Worker Process

Responsibilities:

- Initialize the database connection.
- Register job handlers.
- Poll PostgreSQL for eligible jobs.
- Claim jobs atomically.
- Execute registered handlers.
- Persist execution results.
- Handle graceful shutdown.

### Shared Infrastructure

Both processes will reuse:

- PostgreSQL connection management.
- Job domain models.
- Configuration.
- Shared error definitions.
- Database repositories where appropriate.

## 4. Proposed Project Structure

```text
pulsequeue/
├── cmd/
│   ├── api/
│   │   └── main.go
│   └── worker/
│       └── main.go
│
├── internal/
│   ├── config/
│   ├── database/
│   ├── jobs/
│   │   ├── job.go
│   │   ├── repository.go
│   │   └── postgres_repository.go
│   │
│   └── worker/
│       ├── worker.go
│       ├── registry.go
│       ├── handler.go
│       ├── executor.go
│       ├── poller.go
│       └── errors.go
│
├── migrations/
├── docs/
└── go.mod
```

The structure is a target architecture. Files will be introduced incrementally as implementation responsibilities become necessary.

## 5. Job Execution Lifecycle

The initial execution lifecycle is:

1. A client submits a job.
2. The API validates and stores it in PostgreSQL.
3. The job receives `pending` status.
4. A worker searches for eligible pending jobs.
5. The worker atomically claims a job.
6. The worker increments its attempt counter.
7. The worker resolves the registered handler using the job type.
8. The handler executes.
9. The worker persists the result.
10. The job becomes `completed` or `failed`.

### Initial state transitions

```text
pending
   |
   v
running
   |
   +------> completed
   |
   +------> failed
```

Automatic retry transitions will be implemented in Stage 4.

## 6. Handler Registration Architecture

Developers should register handlers within the worker process.

Proposed registration API:

```go
worker.Register("send_email", SendEmailHandler)
worker.Register("generate_report", GenerateReportHandler)
```

Proposed handler type:

```go
type Handler func(
    ctx context.Context,
    job Job,
) error
```

The context provides cancellation and deadline propagation.

The job argument provides access to the job ID, type, queue and payload.

### Handler registry responsibilities

- Maintain a mapping between job types and handlers.
- Prevent duplicate registrations.
- Reject empty handler names.
- Reject nil handlers.
- Resolve handlers by job type.
- Return a clear error when no handler exists.

Handler registration will be implemented in Stage 3C.

## 7. PostgreSQL Job Claiming Strategy

PostgreSQL will remain the primary job storage and coordination mechanism.

The initial claiming strategy will use:

```sql
SELECT ...
FROM jobs
WHERE
    status = 'pending'
    AND run_at <= NOW()
ORDER BY created_at ASC, id ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;
```

The selection and job ownership update must happen inside a single transaction.

The claim operation will:

1. Begin a transaction.
2. Select an eligible job using row locking.
3. Skip rows locked by other workers.
4. Update the job status to `running`.
5. Increment the attempt count.
6. Record worker ownership and lease information.
7. Commit the transaction.

This prevents multiple workers from simultaneously claiming the same pending row.

### Important implementation requirement

The transaction must remain short.

Do not execute the actual job handler while holding the PostgreSQL row lock.

Claim the job, commit the transaction, and execute the handler afterward.

## 8. Worker Ownership and Leases

A worker must establish ownership when claiming a job.

The following fields are proposed:

| Field | Purpose |
|---|---|
| worker_id | Identifies the worker process |
| locked_at | Records when the job was claimed |
| lease_expires_at | Defines when ownership expires |

These fields will be added through a database migration in Stage 3B.

### Lease concept

A lease represents temporary ownership of a job.

For example:

1. Worker A claims a job.
2. Worker A receives a lease.
3. Worker A executes the job.
4. Worker A renews the lease when necessary.
5. Worker A completes the job and clears ownership.

If the worker crashes, the lease eventually expires.

Another worker can later identify the abandoned job and recover it.

### Important reliability constraint

A lease without renewal is not sufficient for long-running jobs.

Lease renewal and crash recovery must be implemented before relying on the worker engine for production workloads.

The initial lease implementation must also use conditional updates that verify the current worker still owns the job.

## 9. Execution and Failure Handling

Initial behavior:

| Execution result | Job status |
|---|---|
| Handler returns nil | completed |
| Handler returns error | failed |
| Handler panics | failed |
| Handler is not registered | failed |

The execution boundary should recover handler panics, convert them into errors and record useful failure information.

Do not expose sensitive internal details through the public API.

### Retry strategy

Automatic retries are intentionally excluded from the initial worker implementation.

Stage 4 will introduce:

- Retry eligibility.
- Exponential backoff.
- Jitter.
- Maximum attempt enforcement.
- Delayed execution.
- Final failure handling.

## 10. Delivery Semantics and Idempotency

PulseQueue will initially provide at-least-once execution semantics.

Exactly-once execution cannot generally be guaranteed when external side effects and database updates are separate operations.

For example:

1. A handler successfully sends an email.
2. The worker crashes before saving the completed status.
3. The job becomes eligible for recovery.
4. Another worker may execute the handler again.

Therefore, handler implementations should be idempotent wherever possible.

Future documentation should explain idempotency recommendations to developers using the Go SDK.

## 11. Graceful Shutdown

The worker must respond to operating-system shutdown signals.

Expected shutdown behavior:

1. Receive SIGINT or SIGTERM.
2. Stop accepting new jobs.
3. Stop the polling loop.
4. Allow the currently executing handler to finish within a configured grace period.
5. Cancel the execution context when the deadline expires.
6. Persist execution results when possible.
7. Release database resources.
8. Exit cleanly.

A handler must respect its context for cooperative cancellation to work.

Graceful shutdown will be implemented in Stage 3F.

## 12. Initial Worker Configuration

Proposed configuration:

| Setting | Purpose |
|---|---|
| PULSEQUEUE_DATABASE_URL | PostgreSQL connection |
| PULSEQUEUE_WORKER_ID | Worker identity |
| PULSEQUEUE_WORKER_POLL_INTERVAL | Polling interval |
| PULSEQUEUE_WORKER_LEASE_DURATION | Job ownership duration |
| PULSEQUEUE_WORKER_SHUTDOWN_TIMEOUT | Graceful shutdown limit |

Initial concurrency will be one job per worker process.

These settings will be finalized during implementation.

## 13. Testing Strategy

The worker engine requires multiple levels of testing.

### Unit tests

- Handler registration.
- Duplicate handler rejection.
- Unknown handler errors.
- Execution success.
- Handler errors.
- Panic recovery.

### PostgreSQL integration tests

- Successful job claiming.
- Attempt counter increment.
- Ownership assignment.
- Concurrent claim prevention.
- Invalid job state handling.
- Ownership-checked completion.

### Reliability tests

- Worker shutdown.
- Database connection failure.
- Handler execution failure.
- Abandoned lease recovery.
- Concurrent workers.

## 14. Implementation Roadmap

| Stage | Scope |
|---|---|
| 3A | Architecture and design |
| 3B | PostgreSQL job claiming |
| 3C | Handler registration |
| 3D | Worker execution loop |
| 3E | Job completion and failure |
| 3F | Graceful shutdown |
| 3G | Worker integration and concurrency testing |

## 15. Architecture Decisions

### ADR-004: Separate Worker Process

**Decision:** Run workers independently from the API server.

**Reason:** Independent deployment, resource management and scaling.

### ADR-005: PostgreSQL-Based Job Coordination

**Decision:** Use PostgreSQL transactions and `FOR UPDATE SKIP LOCKED`.

**Reason:** Avoid introducing a separate broker in the initial architecture.

### ADR-006: Registered Go Handlers

**Decision:** Use an in-process handler registry.

**Reason:** Type-safe, simple and natural for Go developers.

### ADR-007: Lease-Based Job Ownership

**Decision:** Introduce worker ownership and lease expiration.

**Reason:** Establish a foundation for recovering abandoned jobs.

### ADR-008: At-Least-Once Execution

**Decision:** Design handlers around at-least-once execution.

**Reason:** Database state updates and external side effects cannot generally be committed atomically.

## 16. Future Architecture Extensions

Potential future improvements:

- Concurrent worker execution.
- Worker heartbeats.
- Automatic retries.
- Middleware plugins.
- Job priorities.
- Queue-specific concurrency.
- Execution history.
- Worker monitoring.
- Go SDK.
- Metrics and tracing.

These features should be introduced only when their underlying requirements are understood and validated.

---

**End of Document**

**Status:** Stage 3A – Initial Architecture Proposal
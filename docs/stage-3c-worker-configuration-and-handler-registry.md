# Stage 3C — Worker Configuration and Handler Registry

**Project:** PulseQueue  
**Milestone:** Stage 3 — Background Worker Engine  
**Phase:** 3C — Worker Configuration and Handler Registry  
**Status:** Proposed  
**Prerequisites:** Stage 3A — Worker Engine Architecture; Stage 3B — PostgreSQL Job Claiming  
**Backend:** Go  
**Database:** PostgreSQL

---

# 1. Introduction

## 1.1 Purpose

Stage 3C introduces the configuration and handler registration components required by the PulseQueue worker engine.

Before a worker can execute background jobs, it needs to know:

- Its unique identity.
- Which queues it should process.
- How frequently it should poll for jobs.
- How long a job ownership lease should last.
- How long it should wait during graceful shutdown.
- Which Go handler should execute each job type.

This phase establishes a consistent configuration structure and a handler registry that connects job types to application-defined Go functions.

## 1.2 Background

PulseQueue currently supports persistent job creation, retrieval, cancellation and PostgreSQL job claiming.

Stage 3B introduced the database ownership mechanism that allows workers to claim jobs safely.

However, the worker process still needs two important capabilities:

1. Configuration management.
2. Handler registration and resolution.

For example, when a worker claims a job with type `send_email`, it must identify the corresponding registered Go function.

Stage 3C provides this foundation.

---

# 2. Objectives

The main objectives are:

1. Introduce dedicated worker configuration.
2. Support environment-based configuration.
3. Generate a unique worker identity.
4. Support configurable queue subscriptions.
5. Define polling, lease and shutdown settings.
6. Establish a standard job handler contract.
7. Implement a handler registry.
8. Prevent invalid and duplicate handler registrations.
9. Support handler lookup by job type.
10. Validate configuration and registry behavior through unit tests.
11. Prepare the worker bootstrap for later execution phases.

## 2.1 Out of Scope

The following features are excluded:

- Executing job handlers.
- Polling PostgreSQL for jobs.
- Automatic retries.
- Job completion and failure persistence.
- Full lease renewal implementation.
- Abandoned-job recovery.
- Dynamic runtime handler registration.
- Multiple concurrent job execution.
- Worker monitoring and metrics.

These will be introduced in subsequent stages.

---

# 3. High-Level Architecture

## 3.1 Component Overview

Stage 3C introduces two primary components.

**Worker Configuration**

Responsible for loading, parsing, validating and providing worker-specific settings.

**Handler Registry**

Responsible for registering and resolving Go functions based on job types.

```text
             Worker Process
                   |
          +--------+--------+
          |                 |
    Worker Config      Handler Registry
          |                 |
    Worker Settings    Registered Handlers
          |                 |
          +--------+--------+
                   |
             Worker Engine
              (Stage 3D)
```

The worker engine will consume these components in the next phase.

## 3.2 Example Application Flow

```text
Application Startup
        |
        v
Load Worker Configuration
        |
        v
Validate Configuration
        |
        v
Create Handler Registry
        |
        v
Register Application Handlers
        |
        v
Initialize Worker Engine
        |
        v
Start Processing (Stage 3E)
```

All configuration loading and handler registration should happen before the worker starts accepting jobs.

---

# 4. Worker Configuration

## 4.1 Configuration Structure

Create a dedicated worker configuration structure.

```go
type WorkerConfig struct {
    ID              string
    Queues          []string
    PollInterval    time.Duration
    LeaseDuration   time.Duration
    ShutdownTimeout time.Duration
}
```

### Field Responsibilities

| Field | Description |
|---|---|
| ID | Unique worker process identifier |
| Queues | Queues this worker is allowed to process |
| PollInterval | Time between job polling attempts |
| LeaseDuration | Duration of the job ownership lease |
| ShutdownTimeout | Maximum time allowed for graceful shutdown |

## 4.2 Environment Variables

Worker-specific configuration will use the `PULSEQUEUE_WORKER_` prefix.

| Environment variable | Example | Purpose |
|---|---|---|
| `PULSEQUEUE_WORKER_ID` | `worker-01` | Worker identity |
| `PULSEQUEUE_WORKER_QUEUES` | `default,emails` | Queue subscriptions |
| `PULSEQUEUE_WORKER_POLL_INTERVAL` | `2s` | Polling interval |
| `PULSEQUEUE_WORKER_LEASE_DURATION` | `60s` | Job ownership lease |
| `PULSEQUEUE_WORKER_SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |

The existing shared database configuration remains in the shared configuration package.

Worker-specific settings should not be unnecessarily added to the API server configuration.

## 4.3 Default Configuration

Recommended initial defaults:

| Setting | Default |
|---|---|
| Worker ID | Generated unique ID |
| Queues | `default` |
| Poll interval | `2 seconds` |
| Lease duration | `60 seconds` |
| Shutdown timeout | `30 seconds` |

These are development-friendly starting values, not guaranteed production tuning values.

The final values should be configurable through environment variables.

## 4.4 Worker Identity

Every running worker process needs a unique identity.

Example:

```text
worker-01
worker-02
worker-03
```

If an ID is not explicitly provided, generate a UUID or another collision-resistant process identifier.

Worker IDs must not be treated as permanent identities.

The claim token introduced in Stage 3B remains the authoritative identifier for an individual job claim.

## 4.5 Queue Configuration

A worker may subscribe to one or multiple queues.

Example:

```env
PULSEQUEUE_WORKER_QUEUES=default,emails,reports
```

The configuration loader should:

1. Read the environment variable.
2. Split the value by commas.
3. Trim leading and trailing whitespace.
4. Remove duplicate queue names.
5. Reject empty queue names.
6. Use `default` when the variable is not provided.

For V1, an empty configuration must not silently mean all queues.

## 4.6 Duration Parsing

Go's `time.Duration` supports duration strings.

Examples:

```text
2s
500ms
1m
5m30s
```

Use `time.ParseDuration` rather than manually parsing duration values.

Reject invalid or non-positive values for:

- Poll interval.
- Lease duration.
- Shutdown timeout.

## 4.7 Configuration Loader

A suggested function:

```go
func LoadWorkerConfig() (WorkerConfig, error)
```

Responsibilities:

- Read environment values.
- Apply defaults.
- Parse durations.
- Parse queue subscriptions.
- Generate a worker ID when necessary.
- Validate all configuration values.
- Return a complete configuration or a descriptive error.

The implementation should follow the existing project's configuration and testing conventions.

Dependency injection for environment values may be used if it makes tests easier and avoids changing global process state.

---

# 5. Configuration Validation

Configuration should be validated before creating the worker engine.

## 5.1 Validation Rules

| Field | Validation |
|---|---|
| Worker ID | Must not be empty after trimming |
| Queues | Must contain at least one valid queue |
| Poll interval | Must be greater than zero |
| Lease duration | Must be greater than zero |
| Shutdown timeout | Must be greater than zero |

Additional considerations:

- Queue names should not contain whitespace-only values.
- Duplicate queues should be normalized.
- Explicitly invalid values should return errors rather than silently using defaults.
- Generated IDs should be non-empty.

## 5.2 Lease Configuration

Lease duration must be coordinated with future execution and renewal behavior.

A lease that expires while a handler is still running may allow another worker to recover the same job.

Therefore:

- Lease renewal must be implemented before long-running handlers are supported.
- The renewal interval must leave enough time for database communication and temporary failures.
- Shutdown behavior must account for in-flight jobs.
- Recovery must verify that the original claim is no longer valid.

Stage 3C defines the lease configuration but does not implement lease renewal.

---

# 6. Handler Contract

## 6.1 Purpose

The handler contract defines how application developers implement background job operations.

Each handler receives a job and a context and returns an error if execution fails.

## 6.2 Handler Definition

Recommended contract:

```go
type Handler func(
    ctx context.Context,
    job jobs.Job,
) error
```

The contract uses the existing `jobs.Job` domain model.

### Context

The context provides:

- Cancellation signals.
- Execution deadlines.
- Request-scoped values when appropriate.

Handlers should respect context cancellation.

### Job

The job provides:

- Job ID.
- Job type.
- Queue.
- JSON payload.
- Attempt count.
- Scheduling information.

### Error

The returned error represents unsuccessful execution.

The future execution engine will interpret this result and persist the corresponding job state.

## 6.3 Example Handler

```go
func SendEmailHandler(
    ctx context.Context,
    job jobs.Job,
) error {
    var payload struct {
        To      string `json:"to"`
        Subject string `json:"subject"`
        Body    string `json:"body"`
    }

    if err := json.Unmarshal(job.Payload, &payload); err != nil {
        return fmt.Errorf("decode email payload: %w", err)
    }

    // Perform email operation.

    return nil
}
```

This example demonstrates the handler contract.

Actual external service integrations are outside Stage 3C.

## 6.4 Handler Design Guidelines

Application handlers should:

- Validate their payloads.
- Respect the provided context.
- Return meaningful errors.
- Avoid unnecessarily long blocking operations.
- Be idempotent wherever possible.
- Avoid modifying job lifecycle fields directly.
- Avoid bypassing the worker engine to update job ownership.

Job lifecycle updates belong to the worker execution and repository layers.

---

# 7. Handler Registry

## 7.1 Purpose

The handler registry maps job type names to registered Go functions.

For example:

```text
send_email      -> SendEmailHandler
generate_report -> GenerateReportHandler
process_payment -> ProcessPaymentHandler
```

When a worker receives a job, the registry resolves the correct handler using the job's `Type` field.

## 7.2 Registry Structure

Suggested structure:

```go
type Registry struct {
    handlers map[string]Handler
}
```

The registry stores registered handlers in a Go map.

For V1, handler registration happens during application startup.

After startup, the registry is treated as immutable.

This avoids unnecessary synchronization for the initial implementation.

## 7.3 Registry Creation

Suggested constructor:

```go
func NewRegistry() *Registry {
    return &Registry{
        handlers: make(map[string]Handler),
    }
}
```

The constructor initializes the internal map.

## 7.4 Register Method

Suggested method:

```go
func (r *Registry) Register(
    jobType string,
    handler Handler,
) error
```

Responsibilities:

1. Trim the job type.
2. Reject an empty job type.
3. Reject a nil handler.
4. Reject duplicate registrations.
5. Store the handler.

Example:

```go
registry := worker.NewRegistry()

err := registry.Register(
    "send_email",
    SendEmailHandler,
)
```

## 7.5 Get Method

Suggested method:

```go
func (r *Registry) Get(
    jobType string,
) (Handler, bool)
```

The method returns the handler and a boolean indicating whether it exists.

Example:

```go
handler, ok := registry.Get("send_email")

if !ok {
    return errors.New("handler not registered")
}
```

The registry itself should not decide how unknown job types affect job lifecycle.

That responsibility belongs to the execution engine.

## 7.6 Has Method

Suggested method:

```go
func (r *Registry) Has(jobType string) bool
```

This provides a simple way to check whether a handler has been registered.

It can be useful for startup validation and testing.

---

# 8. Registry Error Handling

The registry should expose clear errors for invalid registrations.

Recommended errors:

```go
var (
    ErrEmptyJobType      = errors.New("job type cannot be empty")
    ErrNilHandler        = errors.New("handler cannot be nil")
    ErrDuplicateHandler  = errors.New("handler already registered")
)
```

Example behavior:

| Scenario | Expected result |
|---|---|
| Empty job type | Registration error |
| Whitespace-only type | Registration error |
| Nil handler | Registration error |
| Duplicate type | Registration error |
| Valid registration | Success |
| Existing handler lookup | Handler returned |
| Unknown handler lookup | Not found |

The registry should not silently overwrite existing handlers.

---

# 9. Worker Bootstrap

## 9.1 Purpose

The worker bootstrap connects configuration, handler registration and the future execution engine.

It will be implemented in:

```text
cmd/worker/main.go
```

## 9.2 Startup Sequence

The worker process should eventually follow this sequence:

1. Load shared application configuration.
2. Load worker configuration.
3. Initialize structured logging.
4. Initialize PostgreSQL connection pool.
5. Verify database readiness.
6. Create handler registry.
7. Register application handlers.
8. Initialize worker engine.
9. Start processing.
10. Handle shutdown signals.
11. Close resources.

## 9.3 Example Bootstrap

Illustrative structure:

```go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    workerCfg, err := worker.LoadWorkerConfig()
    if err != nil {
        log.Fatal(err)
    }

    registry := worker.NewRegistry()

    if err := registry.Register(
        "send_email",
        SendEmailHandler,
    ); err != nil {
        log.Fatal(err)
    }

    // Database initialization and worker engine
    // initialization will be added in later checkpoints.
}
```

This is an architectural illustration, not a complete worker implementation.

Reuse the existing database pool initialization and logging conventions.

Do not duplicate infrastructure initialization unnecessarily.

---

# 10. Suggested Project Structure

The project structure should remain simple and follow the existing internal package organization.

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
│   │   └── config.go
│   │
│   ├── jobs/
│   │   └── ...
│   │
│   ├── worker/
│   │   ├── config.go
│   │   ├── config_test.go
│   │   ├── handler.go
│   │   ├── registry.go
│   │   ├── registry_test.go
│   │   └── errors.go
│   │
│   └── database/
│       └── postgres.go
│
├── migrations/
├── docs/
└── docker-compose.yml
```

Avoid creating additional abstractions unless implementation complexity justifies them.

---

# 11. Testing Strategy

## 11.1 Configuration Tests

Test the following cases:

### Default Values

- Worker ID is generated.
- Default queue is selected.
- Default polling interval is applied.
- Default lease duration is applied.
- Default shutdown timeout is applied.

### Custom Values

- Custom worker ID is accepted.
- Multiple queues are parsed.
- Whitespace is trimmed.
- Duplicate queues are removed.
- Custom durations are parsed correctly.

### Invalid Values

- Invalid duration strings are rejected.
- Zero durations are rejected.
- Negative durations are rejected.
- Empty queue names are rejected.
- Whitespace-only worker IDs are rejected when explicitly supplied.

Use isolated environment inputs in tests to prevent test-order dependencies.

## 11.2 Registry Tests

Test:

- Registry initialization.
- Successful handler registration.
- Successful handler retrieval.
- Unknown handler lookup.
- Empty job type rejection.
- Whitespace-only job type rejection.
- Nil handler rejection.
- Duplicate handler rejection.
- `Has` behavior.

## 11.3 Testing Principles

- Use table-driven tests where appropriate.
- Keep unit tests independent of PostgreSQL.
- Avoid global mutable state.
- Verify both successful and failed operations.
- Ensure error messages or error identities are meaningful.

---

# 12. Implementation Checkpoints

Stage 3C will be implemented incrementally.

| ID | Task | Expected Outcome |
|---|---|---|
| 3C.1 | Implement worker configuration | Environment-based worker settings |
| 3C.2 | Add configuration validation | Invalid settings rejected |
| 3C.3 | Define handler contract | Standard Go handler signature |
| 3C.4 | Implement handler registry | Handler registration and lookup |
| 3C.5 | Implement registry errors | Invalid registrations handled |
| 3C.6 | Add unit tests | Configuration and registry validated |
| 3C.7 | Prepare worker bootstrap | Worker startup foundation |
| 3C.8 | Update documentation | Final design and implementation recorded |
| 3C.9 | Run validation and commit | Stage 3C completed |

---

# 13. Acceptance Criteria

Stage 3C is complete when the following requirements are satisfied.

## Configuration

- [ ] Dedicated worker configuration exists.
- [ ] Environment variables are supported.
- [ ] Worker ID generation works.
- [ ] Queue subscriptions are parsed correctly.
- [ ] Default queue is applied.
- [ ] Duration parsing works.
- [ ] Invalid configuration is rejected.
- [ ] Unit tests cover default, custom and invalid configurations.

## Handler Contract

- [ ] Handler signature is defined.
- [ ] Context is supported.
- [ ] Existing Job domain model is reused.
- [ ] Handler error behavior is documented.

## Registry

- [ ] Registry constructor exists.
- [ ] Handler registration works.
- [ ] Handler lookup works.
- [ ] Handler existence checks work.
- [ ] Empty job types are rejected.
- [ ] Nil handlers are rejected.
- [ ] Duplicate registrations are rejected.
- [ ] Unit tests cover all registry behavior.

## Worker Bootstrap

- [ ] Worker entry point is prepared.
- [ ] Shared configuration is reused.
- [ ] Worker-specific configuration is loaded.
- [ ] Handler registration occurs during startup.
- [ ] Database initialization conventions are reused.

## Validation

- [ ] `go test ./...` passes.
- [ ] `go test -race ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] Documentation is updated.
- [ ] Changes are committed cleanly.

---

# 14. Risks and Technical Considerations

## 14.1 Duplicate Registrations

Allowing duplicate registrations could silently replace an existing handler.

The registry must reject duplicates to prevent unexpected application behavior.

## 14.2 Runtime Registration

Dynamic handler registration is intentionally excluded from V1.

A registry that changes during execution would require synchronization and introduce additional complexity.

If runtime plugins become a requirement, this design can be extended later.

## 14.3 Worker Identity

Worker IDs should identify running processes, not individual job executions.

Claim tokens remain responsible for identifying individual claims.

## 14.4 Lease Configuration

A lease must be long enough for the worker to operate safely and must be renewed for long-running jobs.

Lease renewal and recovery are essential before production workloads are supported.

## 14.5 Configuration Ownership

Shared configuration should remain in the shared configuration package.

Worker-specific settings belong in the worker package.

This avoids coupling the API server to worker-only configuration.

---

# 15. Future Development

Stage 3C prepares the worker engine for execution.

The next phases are:

| Phase | Focus |
|---|---|
| Stage 3D | Job execution engine, panic recovery, ownership-checked result persistence and lease renewal |
| Stage 3E | Continuous polling and job processing |
| Stage 3F | Graceful shutdown and worker lifecycle |
| Stage 4 | Retry strategies, exponential backoff and failure handling |
| Stage 5 | Lease recovery, heartbeat and reliability improvements |
| Stage 6 | Concurrency, scaling and operational monitoring |

Lease renewal and abandoned-job recovery must be implemented before supporting long-running production handlers.

---

# 16. Final Summary

Stage 3C establishes the configuration and handler registration foundation for PulseQueue.

The worker configuration defines how each worker process operates, while the handler registry connects job types to application-defined Go functions.

These components separate infrastructure configuration from application-specific job logic.

After completing Stage 3C, developers will be able to configure a worker process and register the handlers that it will eventually execute.

**Implementation status:** The dedicated worker configuration loader is implemented in
`internal/worker/config.go`, and the handler registry supports registration,
lookup, existence checks, duplicate protection, and validation errors. The worker
bootstrap loads shared database configuration separately from worker-specific
settings before creating the registry and worker engine.

**Stage 3C outcome:** A validated worker configuration system and a predictable, tested handler registry, ready for integration with the PulseQueue execution engine.
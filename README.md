# PulseQueue
A learning-focused, persistent background job engine



## PROJECT MISSION

Build a reliable service that accepts jobs over HTTP, stores them, executes them through Go worker processes, and exposes their status through a modern web dashboard.

### Primary objectives
- Learn Go server-side development beyond basic syntax.
- Understand HTTP, JSON, middleware, interfaces, and error handling.
- Learn PostgreSQL integration and transaction safety.
- Master goroutines, worker pools, cancellation, and concurrency.
- Develop testing, observability, and deployment skills.

  
 ### Initial scope
  `Go standard library`
  `React + Vite`
  `TypeScript`
  `PostgreSQL`
  `Docker Compose`
  
### Success criteria for Milestone 1
- The Go server starts and shuts down correctly.
- A health endpoint returns valid JSON.
- The frontend can request data from the Go API.
- Configuration, logging, and errors have a consistent foundation.
- You can explain the request lifecycle and the role of each main file.

### Runtime configuration

The API binds to `127.0.0.1` by default. Set `PULSEQUEUE_HOST=0.0.0.0`
only when exposing it through an authenticated, protected network boundary.
The frontend uses `/api` by default; set `VITE_API_BASE_URL` when the
production frontend is hosted separately from the API.

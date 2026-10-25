package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	// Call the handler directly
	HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d",
			http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json, got %q", got)
	}

	var response Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status %q, got %q", "ok", response.Status)
	}
}

/*

Why use httptest?
Go provides net/http/httptest so we can test HTTP handlers without starting a real network server.
- httptest.NewRequest creates a test request.
- httptest.NewRecorder captures the response.
- Handler(rec, req) calls the handler directly.
- Assertions verify the HTTP status, content type, and JSON payload.
This is a unit test: it checks one component in isolation. Later, we'll write integration tests that exercise the API together with PostgreSQL.

---------------------------------------------
Run your tests
From the repository root:

go test ./...
go vet ./...

For detailed test output:
go test -v ./...

---------------------------------------------
For race detection:
go test -race ./...

The race detector is particularly valuable later, when we'll have multiple goroutines competing to process jobs. It can detect certain unsynchronized memory accesses, although it cannot prove an entire concurrent system is correct.


*/

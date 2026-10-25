package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/example", nil)
	rec := httptest.NewRecorder()

	handler := LoggingMiddleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		},
	))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rec.Code)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(
		rec,
		http.StatusNotFound,
		"not_found",
		"The requested resource was not found.",
	)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json, got %q", got)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	if response.Error.Code != "not_found" {
		t.Errorf("expected code not_found, got %q", response.Error.Code)
	}

	if !strings.Contains(response.Error.Message, "not found") {
		t.Errorf("unexpected error message: %q", response.Error.Message)
	}
}
